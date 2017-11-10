package testdata

type MyType struct{}

func (m *MyType) MyFunc1() string {
	return ""
}

func (m MyType) MyFunc2() MyIface {
	return nil
}

type MyIface interface {
	Iface() *MyType
}

func MyFuncName() *MyType {
	return nil
}
