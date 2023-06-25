package models

var (
	corpSigningAdapterInstance     corpSigningAdapter
	employeeSigningAdapterInstance employeeSigningAdapter
)

type corpSigningAdapter interface {
	Sign(opt *CorporationSigningCreateOption, linkId string) IModelError
}

type employeeSigningAdapter interface {
	Sign(opt *EmployeeSigning) IModelError
}

func Init(
	cs corpSigningAdapter,
	es employeeSigningAdapter,
) {
	corpSigningAdapterInstance = cs
	employeeSigningAdapterInstance = es
}
