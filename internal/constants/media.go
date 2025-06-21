package constants

type MediaCode string

const (
	MediaCodeDouyin MediaCode = "douyin"
	MediaCodeXhs    MediaCode = "xhs"
)

func (c MediaCode) String() string {
	return string(c)
}
