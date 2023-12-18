package shared

import (
	"xconfwebconfig/db"
	"xconfwebconfig/shared"
)

func GetGenericNamedListOneNonCached(id string) (*shared.GenericNamespacedList, error) {
	instlst, err := db.GetCompressingDataDao().GetOne(db.TABLE_GENERIC_NS_LIST, id)
	if err != nil {
		return nil, err
	}

	if instlst == nil {
		return nil, nil
	}

	lstptr := instlst.(*shared.GenericNamespacedList)

	return lstptr, nil
}
