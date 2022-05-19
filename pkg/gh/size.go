package gh

const (
	labelPrefix = "size/"

	labelXS      = "size/XS"
	labelS       = "size/S"
	labelM       = "size/M"
	labelL       = "size/L"
	labelXL      = "size/XL"
	labelXXL     = "size/XXL"
	labelUnknown = "size/?"
)

const (
	sizeThresholdS   = 10
	sizeThresholdM   = 30
	sizeThresholdL   = 100
	sizeThresholdXL  = 500
	sizeThresholdXXL = 1000
)

type Size int

const (
	SizeXS Size = iota
	SizeS
	SizeM
	SizeL
	SizeXL
	SizeXXL
)

func (s Size) String() string {
	switch s {
	case SizeXS:
		return "XS"
	case SizeS:
		return "S"
	case SizeM:
		return "M"
	case SizeL:
		return "L"
	case SizeXL:
		return "XL"
	case SizeXXL:
		return "XXL"
	default:
		return "Unknown"
	}
}

func (s Size) GetLabel() string {
	switch s {
	case SizeXS:
		return labelXS
	case SizeS:
		return labelS
	case SizeM:
		return labelM
	case SizeL:
		return labelL
	case SizeXL:
		return labelXL
	case SizeXXL:
		return labelXXL
	default:
		return labelUnknown
	}
}

func NewSize(change int) Size {
	switch {
	case change < sizeThresholdS:
		return SizeXS
	case change < sizeThresholdM:
		return SizeS
	case change < sizeThresholdL:
		return SizeM
	case change < sizeThresholdXL:
		return SizeL
	case change < sizeThresholdXXL:
		return SizeXL
	default:
		return SizeXXL
	}
}
