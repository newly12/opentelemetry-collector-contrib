// Copyright  OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mongodbatlasreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver"

import (
	"encoding/json"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver/internal/model"
)

const (
	// Number of log attributes to add to the plog.LogRecordSlice.
	totalLogAttributes = 11

	// Number of resource attributes to add to the plog.ResourceLogs.
	totalResourceAttributes = 4
)

// jsonTimestampLayout for the timestamp format in the plog.Logs structure
const jsonTimestampLayout = "2006-01-02T15:04:05.000-07:00"
const consoleTimestampLayout = "2006-01-02T15:04:05.000-0700"

// Severity mapping of the mongodb atlas logs
var severityMap = map[string]plog.SeverityNumber{
	"F":  plog.SeverityNumberFatal,
	"E":  plog.SeverityNumberError,
	"W":  plog.SeverityNumberWarn,
	"I":  plog.SeverityNumberInfo,
	"D":  plog.SeverityNumberDebug,
	"D1": plog.SeverityNumberDebug,
	"D2": plog.SeverityNumberDebug2,
	"D3": plog.SeverityNumberDebug3,
	"D4": plog.SeverityNumberDebug4,
	"D5": plog.SeverityNumberDebug4,
}

// mongoAuditEventToLogRecord converts model.AuditLog event to plog.LogRecordSlice and adds the resource attributes.
func mongodbAuditEventToLogData(logger *zap.Logger, logs []model.AuditLog, pc ProjectContext, hostname, logName, clusterName, clusterMajorVersion string) plog.Logs {
	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()

	resourceAttrs := rl.Resource().Attributes()
	resourceAttrs.EnsureCapacity(totalResourceAttributes)

	// Attributes related to the object causing the event.
	resourceAttrs.PutString("mongodb_atlas.org", pc.orgName)
	resourceAttrs.PutString("mongodb_atlas.project", pc.Project.Name)
	resourceAttrs.PutString("mongodb_atlas.cluster", clusterName)
	resourceAttrs.PutString("mongodb_atlas.host.name", hostname)

	for _, log := range logs {
		lr := sl.LogRecords().AppendEmpty()
		data, err := json.Marshal(log)
		if err != nil {
			logger.Warn("failed to marshal", zap.Error(err))
		}

		logTsFormat := tsLayout(clusterMajorVersion)
		t, err := time.Parse(logTsFormat, log.Timestamp.Date)
		if err != nil {
			logger.Warn("Time failed to parse correctly", zap.Error(err))
		}

		lr.SetTimestamp(pcommon.NewTimestampFromTime(t))
		lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		// Insert Raw Log message into Body of LogRecord
		lr.Body().SetStringVal(string(data))
		// Since Audit Logs don't have a severity/level
		// Set the "SeverityNumber" and "SeverityText" to INFO
		lr.SetSeverityNumber(plog.SeverityNumberInfo)
		lr.SetSeverityText("INFO")
		attrs := lr.Attributes()
		attrs.EnsureCapacity(totalLogAttributes)
		if log.AuthType != "" {
			attrs.PutString("authtype", log.AuthType)
		}
		attrs.PutString("local.ip", log.Local.IP)
		attrs.PutInt("local.port", int64(log.Local.Port))
		attrs.PutString("remote.ip", log.Remote.IP)
		attrs.PutInt("remote.port", int64(log.Remote.Port))
		attrs.PutString("uuid.binary", log.ID.Binary)
		attrs.PutString("uuid.type", log.ID.Type)
		attrs.PutInt("result", int64(log.Result))
		attrs.PutString("log_name", logName)
		if log.Param.User != "" {
			attrs.PutString("param.user", log.Param.User)
			attrs.PutString("param.database", log.Param.Database)
			attrs.PutString("param.mechanism", log.Param.Mechanism)
		}
	}

	return ld
}

// mongoEventToLogRecord converts model.LogEntry event to plog.LogRecordSlice and adds the resource attributes.
func mongodbEventToLogData(logger *zap.Logger, logs []model.LogEntry, pc ProjectContext, hostname, logName, clusterName, clusterMajorVersion string) plog.Logs {
	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()

	resourceAttrs := rl.Resource().Attributes()
	resourceAttrs.EnsureCapacity(totalResourceAttributes)

	// Attributes related to the object causing the event.
	resourceAttrs.PutString("mongodb_atlas.org", pc.orgName)
	resourceAttrs.PutString("mongodb_atlas.project", pc.Project.Name)
	resourceAttrs.PutString("mongodb_atlas.cluster", clusterName)
	resourceAttrs.PutString("mongodb_atlas.host.name", hostname)

	logTsFormat := tsLayout(clusterMajorVersion)

	for _, log := range logs {
		lr := sl.LogRecords().AppendEmpty()

		rawLog, err := log.RawLog()
		if err != nil {
			logger.Warn("Failed to determine raw log", zap.Error(err))
		}

		t, err := time.Parse(logTsFormat, log.Timestamp.Date)
		if err != nil {
			logger.Warn("Time failed to parse correctly", zap.Error(err))
		}

		lr.SetTimestamp(pcommon.NewTimestampFromTime(t))
		lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		// Insert Raw Log message into Body of LogRecord
		lr.Body().SetStringVal(rawLog)
		// Set the "SeverityNumber" and "SeverityText" if a known type of
		// severity is found.
		if severityNumber, ok := severityMap[log.Severity]; ok {
			lr.SetSeverityNumber(severityNumber)
			lr.SetSeverityText(log.Severity)
		} else {
			logger.Debug("unknown severity type", zap.String("type", log.Severity))
		}
		attrs := lr.Attributes()
		attrs.EnsureCapacity(totalLogAttributes)
		attrs.FromRaw(log.Attributes)
		attrs.PutString("message", log.Message)
		attrs.PutString("component", log.Component)
		attrs.PutString("context", log.Context)
		// log ID is not present on MongoDB 4.2 systems
		if clusterMajorVersion != mongoDBMajorVersion4_2 {
			attrs.PutInt("id", log.ID)
		}
		attrs.PutString("log_name", logName)
		attrs.PutString("raw", rawLog)
	}

	return ld
}

func tsLayout(clusterVersion string) string {
	switch clusterVersion {
	case mongoDBMajorVersion4_2:
		return consoleTimestampLayout
	default:
		return jsonTimestampLayout
	}
}
