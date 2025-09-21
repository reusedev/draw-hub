package image

type AsyncProvider interface {
	Query(ptId int) AsyncQueryResponse
}
