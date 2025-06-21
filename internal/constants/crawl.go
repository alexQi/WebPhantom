package constants

type CrawlerType string

const (
	CrawlerTypeSearch CrawlerType = "search"
	CrawlerTypeMedia  CrawlerType = "media"
	CrawlerTypeUser   CrawlerType = "user"
)

func (c CrawlerType) String() string {
	return string(c)
}
