package goetf

type binaryElement struct {
	// tag type identifier
	tag ExternalTagType
	// if the element is array, slice or map
	complex bool
	// body hold the data of the type
	body []byte
	// items hold the elements for an array or slice
	items []*binaryElement
	// dict hold the pairs for a map
	dict []*binaryElement
}

func newBinaryElement(tag ExternalTagType, body []byte) *binaryElement {
	return &binaryElement{
		tag:     tag,
		body:    body,
		complex: isComplexType(tag),
		items:   make([]*binaryElement, 0),
		dict:    make([]*binaryElement, 0),
	}
}

func (be *binaryElement) append(tag ExternalTagType, elem *binaryElement) {
	switch tag {
	case EttList, EttSmallTuple, EttLargeTuple:
		be.items = append(be.items, elem)
	case EttMap:
		be.dict = append(be.dict, elem)
	default:
		(*be) = *elem
	}
}

func (be *binaryElement) put(tag ExternalTagType, data []byte) {
	elem := newBinaryElement(tag, data)
	switch tag {
	case EttList, EttSmallTuple, EttLargeTuple:
		be.items = append(be.items, elem)
	case EttMap:
		be.dict = append(be.dict, elem)
	default:
		be.body = data
	}
}

func isComplexType(tag ExternalTagType) bool {
	switch tag {
	case EttMap, EttList, EttLargeTuple, EttSmallTuple:
		return true
	}

	return false
}
