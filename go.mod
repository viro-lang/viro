module github.com/marcin-radoszewski/viro

go 1.25.1

require (
	github.com/chzyer/readline v1.5.1
	github.com/ericlagergren/decimal v0.0.0-20240411145413-00de7ca16731
	github.com/webui-dev/go-webui/v2 v2.5.1
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require golang.org/x/sys v0.0.0-20220310020820-b874c991c1a5 // indirect

replace github.com/webui-dev/go-webui/v2 => ./dependencies/go-webui/v2
