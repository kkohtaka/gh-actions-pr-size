package main

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

type size int

const (
	sizeXS size = iota
	sizeS
	sizeM
	sizeL
	sizeXL
	sizeXXL
	sizeUnknown
)

func (s size) String() string {
	switch s {
	case sizeXS:
		return "XS"
	case sizeS:
		return "S"
	case sizeM:
		return "M"
	case sizeL:
		return "L"
	case sizeXL:
		return "XL"
	case sizeXXL:
		return "XXL"
	default:
		return "Unknown"
	}
}

func (s size) getLabel() string {
	switch s {
	case sizeXS:
		return labelXS
	case sizeS:
		return labelS
	case sizeM:
		return labelM
	case sizeL:
		return labelL
	case sizeXL:
		return labelXL
	case sizeXXL:
		return labelXXL
	default:
		return labelUnknown
	}
}

func newSize(change int) size {
	switch {
	case change < sizeThresholdS:
		return sizeXS
	case change < sizeThresholdM:
		return sizeS
	case change < sizeThresholdL:
		return sizeM
	case change < sizeThresholdXL:
		return sizeL
	case change < sizeThresholdXXL:
		return sizeXL
	default:
		return sizeXXL
	}
}
