// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/markbates/pkger"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-grpc/interceptor"
	"github.com/rookie-ninja/rk-grpc/interceptor/context"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"path"
	"strings"
	"time"
)

var (
	Templates = map[string][]byte{}
)

const (
	TvEntryType        = "GrpcTvEntry"
	TvEntryNameDefault = "GrpcTvDefault"
	TvEntryDescription = "Internal RK entry which implements tv web with grpc framework."
)

// Read go TV related template files into memory.
func init() {
	Templates["header"] = readFileFromPkger("/assets/tv/header.tmpl")
	Templates["footer"] = readFileFromPkger("/assets/tv/footer.tmpl")
	Templates["aside"] = readFileFromPkger("/assets/tv/aside.tmpl")
	Templates["head"] = readFileFromPkger("/assets/tv/head.tmpl")
	Templates["svg-sprite"] = readFileFromPkger("/assets/tv/svg-sprite.tmpl")
	Templates["overview"] = readFileFromPkger("/assets/tv/overview.tmpl")
	Templates["apis"] = readFileFromPkger("/assets/tv/apis.tmpl")
	Templates["entries"] = readFileFromPkger("/assets/tv/entries.tmpl")
	Templates["configs"] = readFileFromPkger("/assets/tv/configs.tmpl")
	Templates["certs"] = readFileFromPkger("/assets/tv/certs.tmpl")
	Templates["not-found"] = readFileFromPkger("/assets/tv/not-found.tmpl")
	Templates["internal-error"] = readFileFromPkger("/assets/tv/internal-error.tmpl")
	Templates["os"] = readFileFromPkger("/assets/tv/os.tmpl")
	Templates["env"] = readFileFromPkger("/assets/tv/env.tmpl")
	Templates["prometheus"] = readFileFromPkger("/assets/tv/prometheus.tmpl")
	Templates["deps"] = readFileFromPkger("/assets/tv/deps.tmpl")
	Templates["license"] = readFileFromPkger("/assets/tv/license.tmpl")
	Templates["info"] = readFileFromPkger("/assets/tv/info.tmpl")
	Templates["logs"] = readFileFromPkger("/assets/tv/logs.tmpl")
	Templates["gw-error-mapping"] = readFileFromPkger("/assets/tv/error-mapping.tmpl")
	Templates["git"] = readFileFromPkger("/assets/tv/git.tmpl")
}

// BootConfigTv Bootstrap config of tv.
// 1: Enabled: Enable tv service.
type BootConfigTv struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// TvEntryTvEntry RK TV entry supports web UI for application & process information.
// 1: EntryName: Name of entry.
// 2: EntryType: Type of entry.
// 2: EntryDescription: Description of entry.
// 3: ZapLoggerEntry: ZapLoggerEntry used for logging.
// 4: EventLoggerEntry: EventLoggerEntry used for logging.
// 5: Template: GO template for rendering web UI.
type TvEntry struct {
	EntryName        string                    `json:"entryName" yaml:"entryName"`
	EntryType        string                    `json:"entryType" yaml:"entryType"`
	EntryDescription string                    `json:"entryDescription" yaml:"entryDescription"`
	ZapLoggerEntry   *rkentry.ZapLoggerEntry   `json:"zapLoggerEntry" yaml:"zapLoggerEntry"`
	EventLoggerEntry *rkentry.EventLoggerEntry `json:"eventLoggerEntry" yaml:"eventLoggerEntry"`
	Template         *template.Template        `json:"-" yaml:"-"`
}

// TvEntryOption TV entry option.
type TvEntryOption func(entry *TvEntry)

// WithNameTv Provide name.
func WithNameTv(name string) TvEntryOption {
	return func(entry *TvEntry) {
		entry.EntryName = name
	}
}

// WithEventLoggerEntryTv Provide rkentry.EventLoggerEntry.
func WithEventLoggerEntryTv(eventLoggerEntry *rkentry.EventLoggerEntry) TvEntryOption {
	return func(entry *TvEntry) {
		entry.EventLoggerEntry = eventLoggerEntry
	}
}

// WithZapLoggerEntryTv Provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntryTv(zapLoggerEntry *rkentry.ZapLoggerEntry) TvEntryOption {
	return func(entry *TvEntry) {
		entry.ZapLoggerEntry = zapLoggerEntry
	}
}

// NewTvEntry Create new TV entry with options.
func NewTvEntry(opts ...TvEntryOption) *TvEntry {
	entry := &TvEntry{
		EntryName:        TvEntryNameDefault,
		EntryType:        TvEntryType,
		EntryDescription: TvEntryDescription,
		ZapLoggerEntry:   rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
	}

	for i := range opts {
		opts[i](entry)
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = TvEntryNameDefault
	}

	return entry
}

// AssetsFileHandler Handler which returns js, css, images and html files for TV web UI.
func (entry *TvEntry) AssetsFileHandler(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/rk/v1"), "/")

	if file, err := pkger.Open(path.Join("/boot", p)); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	} else {
		http.ServeContent(w, r, path.Base(p), time.Now(), file)
	}
}

// Bootstrap TV entry.
// Rendering bellow templates.
// 1: head.tmpl
// 2: header.tmpl
// 3: footer.tmpl
// 4: aside.tmpl
// 5: svg-sprite.tmpl
// 6: overview.tmpl
// 7: api.tmpl
// 8: entry.tmpl
// 9: config.tmpl
// 10: cert.tmpl
// 11: os.tmpl
// 12: env.tmpl
// 13: prometheus.tmpl
// 14: log.tmpl
// 15: dep.tmpl
// 16: license.tmpl
// 17: info.tmpl
func (entry *TvEntry) Bootstrap(ctx context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"bootstrap",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	logger := entry.ZapLoggerEntry.GetLogger()

	if raw := ctx.Value(bootstrapEventIdKey); raw != nil {
		event.SetEventId(raw.(string))
		logger = logger.With(zap.String("eventId", event.GetEventId()))
	}

	entry.logBasicInfo(event)

	logger.Info("Bootstrapping TvEntry.", event.ListPayloads()...)

	event.AddPayloads(zap.String("path", "/rk/v1/tv/*item"))

	entry.Template = template.New("rk-tv")

	// Parse templates
	for k, v := range Templates {
		if _, err := entry.Template.Parse(string(v)); err != nil {
			entry.EventLoggerEntry.GetEventHelper().FinishWithError(event, err)
			logger.Error(fmt.Sprintf("Error occurs while parsing %s template.", k))
			rkcommon.ShutdownWithError(err)
		}
	}

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// Interrupt entry.
func (entry *TvEntry) Interrupt(ctx context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"interrupt",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	logger := entry.ZapLoggerEntry.GetLogger()

	if raw := ctx.Value(bootstrapEventIdKey); raw != nil {
		event.SetEventId(raw.(string))
		logger = logger.With(zap.String("eventId", event.GetEventId()))
	}

	entry.logBasicInfo(event)

	defer entry.EventLoggerEntry.GetEventHelper().Finish(event)

	logger.Info("Interrupting TvEntry.", event.ListPayloads()...)
}

// Log basic info into rkquery.Event
func (entry *TvEntry) logBasicInfo(event rkquery.Event) {
	event.AddPayloads(
		zap.String("entryName", entry.EntryName),
		zap.String("entryType", entry.EntryType))
}

// GetName Get name of entry.
func (entry *TvEntry) GetName() string {
	return entry.EntryName
}

// GetType Get type of entry.
func (entry *TvEntry) GetType() string {
	return entry.EntryType
}

// String Stringfy entry.
func (entry *TvEntry) String() string {
	bytesStr, _ := json.Marshal(entry)
	return string(bytesStr)
}

// GetDescription Get description of entry.
func (entry *TvEntry) GetDescription() string {
	return entry.EntryDescription
}

// MarshalJSON Marshal entry
func (entry *TvEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":        entry.EntryName,
		"entryType":        entry.EntryType,
		"entryDescription": entry.EntryDescription,
		"eventLoggerEntry": entry.EventLoggerEntry.GetName(),
		"zapLoggerEntry":   entry.ZapLoggerEntry.GetName(),
	}

	return json.Marshal(&m)
}

// UnmarshalJSON Not supported.
func (entry *TvEntry) UnmarshalJSON([]byte) error {
	return nil
}

// TV Http handler of /rk/v1/tv/*.
func (entry *TvEntry) TV(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/"), "/")

	w.Header().Set("charset", "utf-8")
	w.Header().Set("content-type", "text/html")
	w.Header().Set(rkgrpcctx.RequestIdKey, rkcommon.GenerateRequestId())

	ctx := rkgrpcinter.WrapContextForServer(context.Background())
	rkgrpcinter.AddToServerContextPayload(ctx, rkgrpcinter.RpcEntryNameKey, entry.GetName())

	switch path {
	case "rk/v1/tv", "rk/v1/tv/overview", "rk/v1/tv/application":
		buf := entry.doExecuteTemplate("overview", doReadme(ctx))
		w.Write(buf.Bytes())
	case "rk/v1/tv/apis":
		buf := entry.doExecuteTemplate("apis", doApis(ctx))
		w.Write(buf.Bytes())
	case "rk/v1/tv/entries":
		buf := entry.doExecuteTemplate("entries", doEntries(ctx))
		w.Write(buf.Bytes())
	case "rk/v1/tv/configs":
		buf := entry.doExecuteTemplate("configs", doConfigs(ctx))
		w.Write(buf.Bytes())
	case "rk/v1/tv/certs":
		buf := entry.doExecuteTemplate("certs", doCerts(ctx))
		w.Write(buf.Bytes())
	case "rk/v1/tv/os":
		buf := entry.doExecuteTemplate("os", doSys(ctx))
		w.Write(buf.Bytes())
	case "rk/v1/tv/env":
		buf := entry.doExecuteTemplate("env", doSys(ctx))
		w.Write(buf.Bytes())
	case "rk/v1/tv/prometheus":
		buf := entry.doExecuteTemplate("prometheus", nil)
		w.Write(buf.Bytes())
	case "rk/v1/tv/logs":
		buf := entry.doExecuteTemplate("logs", doLogs(ctx))
		w.Write(buf.Bytes())
	case "rk/v1/tv/deps":
		buf := entry.doExecuteTemplate("deps", doDeps(ctx))
		w.Write(buf.Bytes())
	case "rk/v1/tv/license":
		buf := entry.doExecuteTemplate("license", doLicense(ctx))
		w.Write(buf.Bytes())
	case "rk/v1/tv/info":
		buf := entry.doExecuteTemplate("info", doInfo(ctx))
		w.Write(buf.Bytes())
	case "rk/v1/tv/git":
		buf := entry.doExecuteTemplate("git", doGit(ctx))
		w.Write(buf.Bytes())
	case "rk/v1/tv/gwErrorMapping":
		buf := entry.doExecuteTemplate("gw-error-mapping", doGwErrorMapping(ctx))
		w.Write(buf.Bytes())
	default:
		buf := entry.doExecuteTemplate("not-found", nil)
		w.Write(buf.Bytes())
	}
}

// Execute go template into buffer.
func (entry *TvEntry) doExecuteTemplate(templateName string, data interface{}) *bytes.Buffer {
	buf := new(bytes.Buffer)
	if err := entry.Template.ExecuteTemplate(buf, templateName, data); err != nil {
		entry.ZapLoggerEntry.GetLogger().Warn("Failed to execute template", zap.Error(err))
		buf.Reset()
		entry.Template.ExecuteTemplate(buf, "internal-error", nil)
	}

	return buf
}
