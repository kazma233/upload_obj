package upload

type (
	BedType string
)

type Bed interface {
	Type() BedType
	UploadByPath(filePath string) (string, error)
	UploadByBytes(bs []byte, fname string) (string, error)
	Close()
}

var (
	strategyMap = make(map[BedType]Bed)
)

func CloseAllStrategy() {
	if len(strategyMap) == 0 {
		return
	}

	for _, v := range strategyMap {
		v.Close()
	}

	strategyMap = make(map[BedType]Bed)
}

func RegisterStrategy(bedImpl Bed) {
	if strategyMap[bedImpl.Type()] != nil {
		return
	}

	strategyMap[bedImpl.Type()] = bedImpl
}

func FindStrategy(bedType BedType) Bed {
	return strategyMap[bedType]
}
