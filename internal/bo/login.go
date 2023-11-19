package bo

type UserBO struct {
	Id             int64
	Num            string
	Account        string
	Password       string
	Nickname       string
	Avatar         string
	RegisterSource int32
	Stats          int32
	Token          string
}
