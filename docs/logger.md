# Zap Logger

[Zap logger][uberZap] has ability to attach hooks to the logger.
By using hooks, you can add additional logging destinations.

The following example adds [Application Insights][appInsights] as logging destination:

 ```golang
 import (
	"github.com/Microsoft/ApplicationInsights-Go/appinsights"
	"go.uber.org/zap/zapcore"
)

 func main() {
	client := appinsights.NewTelemetryClient("instrumentation_key")
	logLevelMap := make(map[zapcore.Level]appinsights.SeverityLevel)
	logLevelMap[zapcore.DebugLevel] = appinsights.Verbose
	logLevelMap[zapcore.InfoLevel] = appinsights.Information
	logLevelMap[zapcore.WarnLevel] = appinsights.Warning
	logLevelMap[zapcore.ErrorLevel] = appinsights.Error
	logLevelMap[zapcore.DPanicLevel] = appinsights.Critical
	logLevelMap[zapcore.PanicLevel] = appinsights.Critical
	logLevelMap[zapcore.FatalLevel] = appinsights.Critical

    log, _ := zap.NewDevelopment(zap.Hooks(func(entry zapcore.Entry) error {
	  client.TrackTraceTelemetry(appinsights.NewTraceTelemetry(entry.Message, logLevelMap[entry.Level]))
	  return nil
	}))
}
```
[appInsights]: <https://azure.microsoft.com/en-us/services/application-insights/>
[uberZap]: <https://github.com/uber-go/zap/>
