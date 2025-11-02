package structs

type Uma struct {
	Id   int
	Name string
}

type RelationMember struct {
	A          int
	RelationId int
	UmaId      int
}
