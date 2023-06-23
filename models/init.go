package models

var (
	corpSigningAdapterInstance corpSigningAdapter
)

type corpSigningAdapter interface {
	Sign(opt *CorporationSigningCreateOption, linkId string) IModelError
}

func Init(cs corpSigningAdapter) {
	corpSigningAdapterInstance = cs
}
