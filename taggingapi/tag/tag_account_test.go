package tag

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/rdkcentral/xconfadmin/common"
	taggingapi_config "github.com/rdkcentral/xconfadmin/taggingapi/config"
	"github.com/rdkcentral/xconfadmin/util"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// withTagTypeColumn enables the tag_type column for the duration of a test.
// Off by default so the untyped code path stays the tested default.
func withTagTypeColumn(t *testing.T, enabled bool) {
	t.Helper()
	old := GetTagApiConfig()
	cfg := taggingapi_config.TaggingApiConfig{BatchLimit: 5000, WorkerCount: 20}
	if old != nil {
		cfg = *old
	}
	cfg.TagTypeColumnEnabled = enabled
	SetTagApiConfig(&cfg)
	t.Cleanup(func() { SetTagApiConfig(old) })
}

// --- Account id validation ---

func TestAccountIdValidator(t *testing.T) {
	testCases := []struct {
		name      string
		accountId string
		valid     bool
	}{
		{"19 digit", "2846573900878987927", true},
		{"18 digit", "100498092606141988", true},
		{"12 digit", "100498092606", true},
		{"single digit", "7", true},
		{"empty", "", false},
		{"letters", "not-a-number", false},
		{"hex mac", "AABBCCDDEEFF", false},
		{"colon mac", "AA:BB:CC:DD:EE:FF", false},
		{"leading space", " 123456", false},
		{"negative", "-123456", false},
		{"decimal", "123.456", false},
		{"too long", strings.Repeat("9", 26), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := util.AccountIdValidator(tc.accountId)
			assert.Equal(t, tc.valid, ok)
			assert.Equal(t, tc.valid, err == nil)
			assert.Equal(t, tc.valid, util.IsValidAccountId(tc.accountId))
		})
	}
}

// A 12-digit account id is also a syntactically valid MAC address: the MAC path
// parses it as hex and subtracts 2, silently tagging a different account. This
// is the regression that motivates the separate type.
func TestNormalizeMember_TwelveDigitAccountIdIsNotTreatedAsMac(t *testing.T) {
	const accountId = "100498092606"

	isMac, _ := util.MACAddressValidator(accountId)
	assert.True(t, isMac, "precondition: a 12-digit account id matches the MAC regex")

	assert.NotEqual(t, accountId, ToNormalizedEcm(accountId),
		"precondition: the MAC path corrupts a 12-digit numeric id")

	normalized, err := NormalizeMember(accountId, TagTypeAccount)
	assert.NoError(t, err)
	assert.Equal(t, accountId, normalized, "account ids must pass through unchanged")
}

// The untyped route must keep behaving exactly as it does today, so legacy and
// mac must be byte-identical to ToNormalizedEcm for every input.
func TestNormalizeMember_LegacyAndMacAreUnchanged(t *testing.T) {
	inputs := []string{
		"AA:BB:CC:DD:EE:FF",
		"aa-bb-cc-dd-ee-ff",
		"AABB.CCDD.EEFF",
		"AABBCCDDEEFF",
		"2846573900878987927",
		"100498092606",
		"not-a-mac",
		"  AA:BB:CC:DD:EE:FF  ",
		"",
	}

	for _, input := range inputs {
		t.Run(fmt.Sprintf("input=%q", input), func(t *testing.T) {
			expected := ToNormalizedEcm(input)

			legacy, err := NormalizeMember(input, TagTypeLegacy)
			assert.NoError(t, err)
			assert.Equal(t, expected, legacy)

			mac, err := NormalizeMember(input, TagTypeMac)
			assert.NoError(t, err)
			assert.Equal(t, expected, mac)
		})
	}
}

func TestNormalizeMember_AccountRejectsNonNumeric(t *testing.T) {
	for _, member := range []string{"AA:BB:CC:DD:EE:FF", "abc", "", "12a34"} {
		_, err := NormalizeMember(member, TagTypeAccount)
		assert.Error(t, err, "member %q must be rejected for account tags", member)
		assert.Equal(t, http.StatusBadRequest, xwcommon.GetXconfErrorStatusCode(err))
	}
}

// Whitespace must be stripped before the member reaches Cassandra: the bucket is
// an FNV hash of the stored string, so a padded id would be undeletable.
func TestNormalizeMember_AccountTrimsWhitespace(t *testing.T) {
	normalized, err := NormalizeMember("  2846573900878987927  ", TagTypeAccount)
	assert.NoError(t, err)
	assert.Equal(t, "2846573900878987927", normalized)
	assert.Equal(t, getBucketId("2846573900878987927"), getBucketId(normalized))
}

func TestStoredMemberForm(t *testing.T) {
	// Account tags persist the normalized form so Cassandra and XDAS agree.
	assert.Equal(t, "123", storedMemberForm(" 123 ", "123", TagTypeAccount))
	// Mac tags keep the raw form: GetEcmMacAddress is not idempotent, so storing
	// the normalized value would rebucket every existing member.
	assert.Equal(t, " AA:BB:CC:DD:EE:FF ", storedMemberForm(" AA:BB:CC:DD:EE:FF ", "AABBCCDDEEFD", TagTypeMac))
	assert.Equal(t, "raw", storedMemberForm("raw", "normalized", TagTypeLegacy))
}

func TestValidateTagType(t *testing.T) {
	assert.NoError(t, ValidateTagType(TagTypeLegacy))
	assert.NoError(t, ValidateTagType(TagTypeMac))
	assert.NoError(t, ValidateTagType(TagTypeAccount))

	err := ValidateTagType("device")
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, xwcommon.GetXconfErrorStatusCode(err))
}

// --- Tag type resolution and uniqueness ---

func TestResolveTagType(t *testing.T) {
	testCases := []struct {
		name     string
		rowTypes []string
		expected string
	}{
		{"no rows", nil, TagTypeLegacy},
		{"all legacy", []string{"", "", ""}, TagTypeLegacy},
		{"all account", []string{"account", "account"}, TagTypeAccount},
		// Mixed rows only arise from two concurrent first-ever adds of the same
		// id; account wins so the type cannot flap.
		{"mixed", []string{"", "account", ""}, TagTypeAccount},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, resolveTagType(tc.rowTypes))
		})
	}
}

func TestEnsureTagTypeCompatible(t *testing.T) {
	testCases := []struct {
		name          string
		existingType  string
		requestedType string
		expectStatus  int // 0 means no error
	}{
		// An existing tag with untyped rows is a mac tag that predates the column.
		{"legacy tag as account", TagTypeLegacy, TagTypeAccount, http.StatusConflict},
		{"legacy tag as mac", TagTypeLegacy, TagTypeMac, 0},
		{"account tag as account", TagTypeAccount, TagTypeAccount, 0},
		{"account tag as mac", TagTypeAccount, TagTypeMac, http.StatusConflict},
		// Legacy requests are never a conflict here: effectiveWriteTagType
		// resolves them to the stored type before any store is touched.
		{"account tag via legacy route", TagTypeAccount, TagTypeLegacy, 0},
		{"mac tag via legacy route", TagTypeLegacy, TagTypeLegacy, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestEnvironment()
			withTagTypeColumn(t, true)
			withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
				if isMetadataQuery(query) {
					return []map[string]any{
						{"bucket_id": 1, "tag_type": tc.existingType},
					}, nil
				}
				return []map[string]any{}, nil
			})

			err := ensureTagTypeCompatible("some-tag", tc.requestedType)
			if tc.expectStatus == 0 {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
			assert.Equal(t, tc.expectStatus, xwcommon.GetXconfErrorStatusCode(err))
		})
	}
}

// A tag with no metadata rows does not exist yet and may be claimed by any type.
// Deciding on the stored type alone conflates "absent" with "legacy" — both are
// the empty string — and rejects the first write of every new account tag.
func TestEnsureTagTypeCompatible_UnknownTagIsClaimable(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return []map[string]any{}, nil
	})

	assert.NoError(t, ensureTagTypeCompatible("brand-new", TagTypeMac))
	assert.NoError(t, ensureTagTypeCompatible("brand-new", TagTypeAccount))
}

// Existence is the bucket count, not the type: untyped rows are an existing mac
// tag from before the column and must still block an account claim, even though
// the type reads back identical to an absent tag's.
func TestEnsureTagTypeCompatible_ExistingUntypedTagBlocksAccountClaim(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 42, "tag_type": TagTypeLegacy}}, nil
		}
		return []map[string]any{}, nil
	})

	err := ensureTagTypeCompatible("pre-existing-mac-tag", TagTypeAccount)
	assert.Error(t, err)
	assert.Equal(t, http.StatusConflict, xwcommon.GetXconfErrorStatusCode(err))
}

// A DB error must not read as "no conflict".
func TestEnsureTagTypeCompatible_DbErrorFailsClosed(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return nil, fmt.Errorf("cassandra unavailable")
	})

	err := ensureTagTypeCompatible("some-tag", TagTypeAccount)
	assert.Error(t, err)
	assert.NotEqual(t, http.StatusConflict, xwcommon.GetXconfErrorStatusCode(err),
		"a DB failure is not a type conflict")
}

// The comparison half performs no database access: the delete handler passes the
// type its own getTagMeta already resolved, and a hidden re-read would double the
// metadata cost of every typed delete.
func TestCheckTagTypeCompatible_NoDatabaseRead(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		t.Error("no database query expected from the comparison half")
		return nil, nil
	})

	assert.NoError(t, checkTagTypeCompatible("t", TagTypeAccount, TagTypeAccount))
	assert.NoError(t, checkTagTypeCompatible("t", TagTypeLegacy, TagTypeMac))

	err := checkTagTypeCompatible("t", TagTypeLegacy, TagTypeAccount)
	assert.Equal(t, http.StatusConflict, xwcommon.GetXconfErrorStatusCode(err))
	err = checkTagTypeCompatible("t", TagTypeAccount, TagTypeMac)
	assert.Equal(t, http.StatusConflict, xwcommon.GetXconfErrorStatusCode(err))
}

// With the column disabled the check is inert, so the binary can run against a
// cluster that has not had the ALTER applied.
func TestEnsureTagTypeCompatible_NoOpWhenColumnDisabled(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, false)
	withMockDbClient(t, &mockDbClient{}) // panics if any query is issued

	assert.NoError(t, ensureTagTypeCompatible("some-tag", TagTypeAccount))
}

// --- Persistence ---

func TestAddMembersToBucket_WritesTagTypeOnlyOnMetadata(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var captured *mockBatch
	withMockDbClient(t, &mockDbClient{
		execBatchFn: func(batch *mockBatch) error {
			captured = batch
			return nil
		},
	})

	err := addMembersToBucket("acct-tag", 7, []string{"2846573900878987927"}, "123", TagTypeAccount)
	assert.NoError(t, err)
	assert.NotNil(t, captured)

	// The member insert stays untyped: a tag_type cell per member would cost a
	// text column on the hottest write path with no reader.
	memberArgs := captured.argsFor(QueryAddMemberBucketed)
	assert.NotNil(t, memberArgs)
	assert.Len(t, memberArgs, 4)
	for _, stmt := range captured.statements {
		assert.NotContains(t, stmt, `"TagMembersBucketed" (tag_id, bucket_id, member, created, tag_type)`)
	}

	metadataArgs := captured.argsFor(QueryAddBucketMetadataTyped)
	assert.NotNil(t, metadataArgs, "metadata insert must use the typed statement")
	assert.Len(t, metadataArgs, 3)
	assert.Equal(t, TagTypeAccount, metadataArgs[2])
}

// A mac-typed add stores tag_type as legacy (""). Mac and legacy are one
// equivalence class, so a persisted "mac" would be a third stored value that
// every reader collapses anyway, with raw rows disagreeing with what reads report.
func TestAddMembersToBucket_MacTypeStoredAsLegacy(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var captured *mockBatch
	withMockDbClient(t, &mockDbClient{
		execBatchFn: func(batch *mockBatch) error {
			captured = batch
			return nil
		},
	})

	err := addMembersToBucket("mac-tag", 7, []string{"AABBCCDDEEFF"}, "123", TagTypeMac)
	assert.NoError(t, err)
	assert.NotNil(t, captured)

	metadataArgs := captured.argsFor(QueryAddBucketMetadataTyped)
	assert.NotNil(t, metadataArgs, "metadata insert must use the typed statement")
	assert.Len(t, metadataArgs, 3)
	assert.Equal(t, TagTypeLegacy, metadataArgs[2])
}

func TestAddMembersToBucket_UsesUntypedStatementWhenColumnDisabled(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, false)

	var captured *mockBatch
	withMockDbClient(t, &mockDbClient{
		execBatchFn: func(batch *mockBatch) error {
			captured = batch
			return nil
		},
	})

	err := addMembersToBucket("mac-tag", 7, []string{"AABBCCDDEEFF"}, "123", TagTypeLegacy)
	assert.NoError(t, err)
	assert.NotNil(t, captured)

	// Naming a column the cluster lacks fails the whole batch, taking every tag
	// add down with it — not just account ones.
	assert.Nil(t, captured.argsFor(QueryAddBucketMetadataTyped))
	assert.NotNil(t, captured.argsFor(QueryAddBucketMetadata))
}

// --- Listing and filtering ---

func TestGetAllTagIds_FiltersByType(t *testing.T) {
	rows := []map[string]any{
		{"tag_id": "acct-tag", "tag_type": "account"},
		{"tag_id": "acct-tag", "tag_type": "account"}, // second bucket
		{"tag_id": "mac-tag", "tag_type": ""},
		{"tag_id": "legacy-tag"}, // column absent entirely
	}

	testCases := []struct {
		name     string
		filter   string
		expected []string
	}{
		{"no filter returns all", TagTypeLegacy, []string{"acct-tag", "mac-tag", "legacy-tag"}},
		{"account", TagTypeAccount, []string{"acct-tag"}},
		// A mac filter must include legacy rows: legacy tags are mac tags that
		// predate the column.
		{"mac includes legacy", TagTypeMac, []string{"mac-tag", "legacy-tag"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestEnvironment()
			withTagTypeColumn(t, true)

			var seenQuery string
			withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
				seenQuery = query
				return rows, nil
			})

			tagIds, err := GetAllTagIds(tc.filter)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tc.expected, tagIds)

			// Filtering happens in Go: a secondary index on a two-valued column
			// would hotspot two nodes, and ALLOW FILTERING is worse than the scan.
			assert.NotContains(t, seenQuery, "WHERE")
			assert.NotContains(t, strings.ToUpper(seenQuery), "ALLOW FILTERING")
		})
	}
}

// One row exists per (tag_id, bucket_id), so a tag spread over many buckets must
// still appear once.
func TestGetAllTagIds_DedupesAcrossBuckets(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	rows := make([]map[string]any, 0, 100)
	for i := 0; i < 100; i++ {
		rows = append(rows, map[string]any{"tag_id": "spread-tag", "tag_type": "account"})
	}
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return rows, nil
	})

	tagIds, err := GetAllTagIds(TagTypeAccount)
	assert.NoError(t, err)
	assert.Equal(t, []string{"spread-tag"}, tagIds)
}

// Tag ids are stored unprefixed, so the listing must not strip a leading "t_"
// from a tag legitimately named that way.
func TestGetAllTagIds_DoesNotStripTagPrefix(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, false)
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return []map[string]any{
			{"tag_id": "t_beta"},
			{"tag_id": "beta"},
		}, nil
	})

	tagIds, err := GetAllTagIds(TagTypeLegacy)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"t_beta", "beta"}, tagIds,
		"a tag named t_beta must not collapse onto beta")
}

// --- Shared keyspace ---

// Account and mac members share the device XDAS keyspace; an account write must
// reach it with the id unmodified (numeric validation, no ECM shift).
func TestAddMembers_AccountTagUsesSharedKeyspace(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var paths []string
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})

	// Empty DB: "acct-tag" has no owner yet, the ordinary way one gets created.
	withMockDbClient(t, &mockDbClient{
		queryFunc: func(query string, params ...string) ([]map[string]any, error) {
			return []map[string]any{}, nil
		},
	})

	stats, err := AddMembersWithXdas("acct-tag", []string{"2846573900878987927"}, "", TagTypeAccount)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.XdasOk)

	assert.Len(t, paths, 1)
	assert.True(t, strings.HasPrefix(paths[0], "/ft/"), "got %q, want the shared keyspace", paths[0])
	assert.Contains(t, paths[0], "2846573900878987927")
}

func TestAddMembers_LegacyTagStillUsesDeviceKeyspace(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var paths []string
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})
	withMockDbClient(t, &mockDbClient{
		queryFunc: func(query string, params ...string) ([]map[string]any, error) {
			return []map[string]any{}, nil
		},
	})

	stats, err := AddMembersWithXdas("mac-tag", []string{"AA:BB:CC:DD:EE:FF"}, "", TagTypeLegacy)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.XdasOk)

	assert.Len(t, paths, 1)
	assert.True(t, strings.HasPrefix(paths[0], "/ft/"), "got %q, want the device keyspace", paths[0])
}

// Deleting an account tag through the untyped route must resolve the account
// type from storage: a 12-digit id would otherwise go through the MAC path and
// be ECM-shifted to a different XDAS key, orphaning the real entry.
func TestDeleteTag_LegacyRouteResolvesAccountTypeFromStorage(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var deletePaths []string
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deletePaths = append(deletePaths, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})

	var memberFetches atomic.Int32
	withMockDbClient(t, &mockDbClient{
		queryFunc: func(query string, params ...string) ([]map[string]any, error) {
			if isMetadataQuery(query) {
				return []map[string]any{{"bucket_id": 3, "tag_type": TagTypeAccount}}, nil
			}
			// One page of members, then empty so the delete loop terminates.
			if memberFetches.Add(1) == 1 {
				return []map[string]any{{"member": "2846573900878987927"}}, nil
			}
			return []map[string]any{}, nil
		},
	})

	err := DeleteTag("acct-tag", "audit-id")
	assert.NoError(t, err)

	assert.NotEmpty(t, deletePaths, "expected XDAS deletes to be issued")
	for _, path := range deletePaths {
		assert.Contains(t, path, "2846573900878987927",
			"got %q, want the stored account id deleted unmodified", path)
	}
}

// --- Untyped writes adopt the stored type ---

// An untyped write to an account tag must execute as an account write. Taking
// the type from the route would run its members through the MAC normalizer —
// ECM-shifting 12-digit ids under a different XDAS key than the Cassandra rows —
// a mixture DeleteTag can never remove.
func TestAddMembers_UntypedWriteToAccountTagAdoptsAccountType(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var paths []string
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})

	var captured *mockBatch
	withMockDbClient(t, &mockDbClient{
		queryFunc: func(query string, params ...string) ([]map[string]any, error) {
			if isMetadataQuery(query) {
				return []map[string]any{{"bucket_id": 3, "tag_type": TagTypeAccount}}, nil
			}
			return []map[string]any{}, nil
		},
		execBatchFn: func(batch *mockBatch) error {
			captured = batch
			return nil
		},
	})

	stats, err := AddMembersWithXdas("acct-tag", []string{"2846573900878987927"}, "", TagTypeLegacy)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.XdasOk)
	assert.Equal(t, TagTypeAccount, stats.TagType)

	assert.Len(t, paths, 1)
	assert.Contains(t, paths[0], "2846573900878987927",
		"got %q, want the account id written unmodified", paths[0])

	// The rows keep the account type, so the tag cannot decay toward legacy.
	assert.NotNil(t, captured)
	metadataArgs := captured.argsFor(QueryAddBucketMetadataTyped)
	assert.NotNil(t, metadataArgs)
	assert.Equal(t, TagTypeAccount, metadataArgs[2])
}

// MAC members on an untyped write to an account tag fail account normalization
// and must not reach XDAS at all.
func TestAddMembers_UntypedWriteOfMacsToAccountTagIssuesNoXdasCalls(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var requests atomic.Int32
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 3, "tag_type": TagTypeAccount}}, nil
		}
		return []map[string]any{}, nil
	})

	stats, err := AddMembersWithXdas("acct-tag", []string{"AA:BB:CC:DD:EE:FF"}, "", TagTypeLegacy)
	assert.Error(t, err)
	assert.Equal(t, 0, stats.XdasOk)
	assert.Equal(t, 1, stats.XdasFail)
	assert.Equal(t, int32(0), requests.Load(),
		"MAC members of an account tag must not reach XDAS")
}

func TestRemoveMembers_UntypedRemoveFromAccountTagAdoptsAccountType(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var deletePaths []string
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deletePaths = append(deletePaths, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 3, "tag_type": TagTypeAccount}}, nil
		}
		return []map[string]any{}, nil
	})

	stats, err := RemoveMembersWithXdas("acct-tag", []string{"2846573900878987927"}, TagTypeLegacy)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.XdasOk)
	assert.Equal(t, TagTypeAccount, stats.TagType)

	assert.Len(t, deletePaths, 1)
	assert.Contains(t, deletePaths[0], "2846573900878987927",
		"got %q, want the account id removed unmodified", deletePaths[0])
}

// The stored type of a mac/legacy tag resolves to legacy, so untyped writes to
// existing tags keep today's behavior byte-for-byte.
func TestAddMembers_UntypedWriteToExistingLegacyTagKeepsLegacyBehavior(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var paths []string
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			return []map[string]any{{"bucket_id": 1, "tag_type": TagTypeLegacy}}, nil
		}
		return []map[string]any{}, nil
	})

	stats, err := AddMembersWithXdas("mac-tag", []string{"AA:BB:CC:DD:EE:FF"}, "", TagTypeLegacy)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.XdasOk)

	assert.Len(t, paths, 1)
	assert.True(t, strings.HasPrefix(paths[0], "/ft/"), "got %q, want the device keyspace", paths[0])
}

// An unreadable stored type must not default the write to legacy — account
// members would silently go through the MAC normalizer.
func TestAddMembers_UntypedWriteFailsClosedOnMetadataError(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var requests atomic.Int32
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return nil, fmt.Errorf("cassandra unavailable")
	})

	stats, err := AddMembersWithXdas("some-tag", []string{"AA:BB:CC:DD:EE:FF"}, "", TagTypeLegacy)
	assert.Error(t, err)
	assert.Equal(t, 0, stats.XdasOk)
	assert.Equal(t, int32(0), requests.Load())
}

// --- Fail-loud behavior ---

// A blank add template must fail the write rather than fmt.Sprintf its way to a
// garbage URL; a member "stored" that way is unrecoverable.
func TestAccountWrite_FailsClosedWithoutTemplate(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var requests atomic.Int32
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	// Blank the template withMockXdasSync just set; restore it for later tests.
	GetGroupServiceSyncConnector().SetAddGroupMemberTemplate("")
	t.Cleanup(func() { GetGroupServiceSyncConnector().SetAddGroupMemberTemplate("%s/ft/%s") })

	withMockDbClient(t, &mockDbClient{
		queryFunc: func(query string, params ...string) ([]map[string]any, error) {
			if isMetadataQuery(query) {
				return []map[string]any{{"bucket_id": 1, "tag_type": TagTypeAccount}}, nil
			}
			return []map[string]any{}, nil
		},
	})

	stats, err := AddMembersWithXdas("acct-tag", []string{"2846573900878987927"}, "", TagTypeAccount)
	assert.Error(t, err)
	assert.Equal(t, 0, stats.XdasOk)
	assert.Equal(t, int32(0), requests.Load(), "no request may be issued with a blank template")
}

// An invalid member must be rejected before any XDAS call.
func TestAccountWrite_InvalidMemberIssuesNoXdasCall(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var requests atomic.Int32
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	withMockDbClient(t, &mockDbClient{
		queryFunc: func(query string, params ...string) ([]map[string]any, error) {
			if isMetadataQuery(query) {
				return []map[string]any{{"bucket_id": 1, "tag_type": TagTypeAccount}}, nil
			}
			return []map[string]any{}, nil
		},
	})

	stats, err := AddMembersWithXdas("acct-tag", []string{"AA:BB:CC:DD:EE:FF"}, "", TagTypeAccount)
	assert.Error(t, err)
	assert.Equal(t, 0, stats.XdasOk)
	assert.Equal(t, 1, stats.XdasFail)
	assert.Equal(t, int32(0), requests.Load())
}

// --- Reserved tag ids ---

func TestValidateTagId_RejectsReservedIds(t *testing.T) {
	for _, id := range []string{"mac", "account", "members", "MAC", "Account"} {
		err := validateTagId(id)
		assert.Error(t, err, "tag id %q must be reserved", id)
		assert.Equal(t, http.StatusBadRequest, xwcommon.GetXconfErrorStatusCode(err))
	}

	for _, id := range []string{"promo2026", "x1:product:customization:cox:btrh0029", "macaroni", "t_beta"} {
		assert.NoError(t, validateTagId(id), "tag id %q must be allowed", id)
	}
}

func TestAddMembersToTagHandler_RejectsReservedTagId(t *testing.T) {
	setupTestEnvironment()

	req := httptest.NewRequest(http.MethodPut, "/taggingService/tags/account/members", nil)
	req = mux.SetURLVars(req, map[string]string{common.Tag: "account"})
	recorder := httptest.NewRecorder()
	xw := &xwhttp.XResponseWriter{ResponseWriter: recorder}
	xw.SetBody(`["AA:BB:CC:DD:EE:FF"]`)

	AddMembersToTagHandler(xw, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "reserved")
}

func TestGetTagTypeFromRequest(t *testing.T) {
	testCases := []struct {
		name       string
		vars       map[string]string
		query      string
		columnOn   bool
		expected   string
		expectErr  bool
		expectCode int
	}{
		{"absent means legacy", nil, "", true, TagTypeLegacy, false, 0},
		{"path var account", map[string]string{common.TagType: TagTypeAccount}, "", true, TagTypeAccount, false, 0},
		{"path var mac", map[string]string{common.TagType: TagTypeMac}, "", true, TagTypeMac, false, 0},
		{"query fallback", nil, "?tagType=account", true, TagTypeAccount, false, 0},
		{"unknown type", map[string]string{common.TagType: "device"}, "", true, "", true, http.StatusBadRequest},
		{"unknown query type", nil, "?tagType=device", true, "", true, http.StatusBadRequest},

		// The account type is unusable without the column, so it is refused
		// rather than silently degraded to legacy.
		{"account rejected when column disabled", map[string]string{common.TagType: TagTypeAccount}, "", false, "", true, http.StatusServiceUnavailable},
		{"account query rejected when column disabled", nil, "?tagType=account", false, "", true, http.StatusServiceUnavailable},

		// The untyped and mac routes must keep working without the ALTER.
		{"legacy allowed when column disabled", nil, "", false, TagTypeLegacy, false, 0},
		{"mac allowed when column disabled", map[string]string{common.TagType: TagTypeMac}, "", false, TagTypeMac, false, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestEnvironment()
			withTagTypeColumn(t, tc.columnOn)

			req := httptest.NewRequest(http.MethodGet, "/taggingService/tags"+tc.query, nil)
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}

			tagType, err := getTagTypeFromRequest(req)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expectCode, xwcommon.GetXconfErrorStatusCode(err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, tagType)
		})
	}
}

// --- Account routes are refused while the column is disabled ---

// Accepting the write would answer 200 while storing an untyped metadata row, so
// the tag reads back as legacy forever — invisible to the account listing — with
// its members already in XDAS.
func TestAddMembersToTagHandler_RejectsAccountWhenColumnDisabled(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, false)
	withMockDbClient(t, &mockDbClient{}) // panics if any query is issued

	var xdasRequests atomic.Int32
	withMockXdasSync(t, func(w http.ResponseWriter, r *http.Request) {
		xdasRequests.Add(1)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/taggingService/tags/account/acct-tag/members", nil)
	req = mux.SetURLVars(req, map[string]string{common.Tag: "acct-tag", common.TagType: TagTypeAccount})
	recorder := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(recorder)
	xw.SetBody(`["2846573900878987927"]`)

	AddMembersToTagHandler(xw, req)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "tag_type_column_enabled")
	assert.Equal(t, int32(0), xdasRequests.Load(), "nothing may reach XDAS")
}

// The listing can only ever come back empty without the column, which reads as
// "no account tags exist" rather than "account tags are not available here".
func TestGetAllTagsHandler_RejectsAccountWhenColumnDisabled(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, false)
	withMockDbClient(t, &mockDbClient{}) // panics if any query is issued

	req := httptest.NewRequest(http.MethodGet, "/taggingService/tags/account", nil)
	req = mux.SetURLVars(req, map[string]string{common.TagType: TagTypeAccount})
	recorder := httptest.NewRecorder()

	GetAllTagsHandler(xwhttp.NewXResponseWriter(recorder), req)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

// The mac and untyped routes must keep serving on a cluster without the ALTER.
func TestGetAllTagsHandler_MacRouteStillServesWhenColumnDisabled(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, false)
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		return []map[string]any{{"tag_id": "beta"}}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/taggingService/tags/mac", nil)
	req = mux.SetURLVars(req, map[string]string{common.TagType: TagTypeMac})
	recorder := httptest.NewRecorder()

	GetAllTagsHandler(xwhttp.NewXResponseWriter(recorder), req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "beta")
}

// --- Typed reads ---

// withStoredTagOfType serves a one-bucket metadata partition of storedType plus
// one member, counting member queries separately: a tag scoped away by the route
// type must be rejected off the metadata read alone.
func withStoredTagOfType(t *testing.T, storedType string, memberQueries *atomic.Int32) {
	t.Helper()
	withMockDb(t, func(query string, params ...string) ([]map[string]any, error) {
		if isMetadataQuery(query) {
			row := map[string]any{"bucket_id": 7}
			// A legacy tag has no tag_type cell, like rows written before it.
			if storedType != TagTypeLegacy {
				row["tag_type"] = storedType
			}
			return []map[string]any{row}, nil
		}
		memberQueries.Add(1)
		return []map[string]any{{"member": "2846573900878987927"}}, nil
	})
}

// A typed read must not serve a tag stored under another type. /tags/mac/{id}
// used to return an account tag in full while the /tags/mac listing omitted it
// and DELETE on the same path 409'd — one tag, three answers.
func TestTypedReads_ScopeByStoredTagType(t *testing.T) {
	readers := []struct {
		name string
		read func(tagId string, tagType string) (int, error)
	}{
		{"GetTagById", func(tagId string, tagType string) (int, error) {
			members, _, _, err := GetTagById(tagId, tagType)
			return len(members), err
		}},
		{"GetMembersNonPaginated", func(tagId string, tagType string) (int, error) {
			members, _, _, err := GetMembersNonPaginated(tagId, tagType)
			return len(members), err
		}},
		{"GetMembersPaginated", func(tagId string, tagType string) (int, error) {
			resp, _, err := GetMembersPaginated(tagId, 10, "", tagType)
			if resp == nil {
				return 0, err
			}
			return len(resp.Data), err
		}},
	}

	testCases := []struct {
		name       string
		storedType string
		routeType  string
		visible    bool
	}{
		{"mac route hides account tag", TagTypeAccount, TagTypeMac, false},
		{"account route hides legacy tag", TagTypeLegacy, TagTypeAccount, false},
		{"account route serves account tag", TagTypeAccount, TagTypeAccount, true},
		// Legacy tags predate the column, so the mac route owns them.
		{"mac route serves legacy tag", TagTypeLegacy, TagTypeMac, true},
		// The untyped routes stay type-blind, byte for byte as before.
		{"untyped route serves account tag", TagTypeAccount, TagTypeLegacy, true},
		{"untyped route serves legacy tag", TagTypeLegacy, TagTypeLegacy, true},
	}

	for _, reader := range readers {
		for _, tc := range testCases {
			t.Run(reader.name+"/"+tc.name, func(t *testing.T) {
				setupTestEnvironment()
				withTagTypeColumn(t, true)

				var memberQueries atomic.Int32
				withStoredTagOfType(t, tc.storedType, &memberQueries)

				count, err := reader.read("some-tag", tc.routeType)
				if tc.visible {
					assert.NoError(t, err)
					assert.Equal(t, 1, count)
					return
				}

				assert.Error(t, err)
				assert.Equal(t, http.StatusNotFound, xwcommon.GetXconfErrorStatusCode(err))
				assert.Equal(t, int32(0), memberQueries.Load(),
					"a tag scoped away by its type must not have its buckets read")
			})
		}
	}
}

// Without the column every stored type resolves to legacy, so the mac route
// keeps serving on a cluster without the ALTER.
func TestTypedReads_MacRouteStillServesWhenColumnDisabled(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, false)

	var memberQueries atomic.Int32
	withStoredTagOfType(t, TagTypeLegacy, &memberQueries)

	members, _, _, err := GetTagById("some-tag", TagTypeMac)
	assert.NoError(t, err)
	assert.Len(t, members, 1)
}

// End to end on the reported request: GET /taggingService/tags/mac/{accountTag}.
func TestGetTagByIdHandler_MacRouteDoesNotServeAccountTag(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var memberQueries atomic.Int32
	withStoredTagOfType(t, TagTypeAccount, &memberQueries)

	req := httptest.NewRequest(http.MethodGet, "/taggingService/tags/mac/acct-tag", nil)
	req = mux.SetURLVars(req, map[string]string{common.Tag: "acct-tag", common.TagType: TagTypeMac})
	recorder := httptest.NewRecorder()

	GetTagByIdHandler(xwhttp.NewXResponseWriter(recorder), req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, int32(0), memberQueries.Load())
}

func TestGetTagMembersHandler_AccountRouteServesAccountTag(t *testing.T) {
	setupTestEnvironment()
	withTagTypeColumn(t, true)

	var memberQueries atomic.Int32
	withStoredTagOfType(t, TagTypeAccount, &memberQueries)

	req := httptest.NewRequest(http.MethodGet, "/taggingService/tags/account/acct-tag/members", nil)
	req = mux.SetURLVars(req, map[string]string{common.Tag: "acct-tag", common.TagType: TagTypeAccount})
	recorder := httptest.NewRecorder()

	GetTagMembersHandler(xwhttp.NewXResponseWriter(recorder), req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "2846573900878987927")
}
