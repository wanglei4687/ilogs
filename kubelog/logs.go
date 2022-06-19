package kubelog

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
)

var logFlushFreq = flag.Duration("log-flush-frequency", 5*time.Second, "Maximum number of seconds between log flushes")
var Log = klogr.New().WithName(os.Getenv("LOGNAME"))

type GlogWriter struct{}

func (writers GlogWriter) Write(data []byte) (n int, err error) {
	klog.Info(string(data))

	return len(data), nil
}

func InitLog(fs *flag.FlagSet) {
	if fs == nil {
		fs = flag.CommandLine
	}

	klog.InitFlags(fs)

	_ = fs.Set("log_to_stderr", "true")

	log.SetOutput(GlogWriter{})
	log.SetFlags(0)

	go wait.Until(klog.Flush, *logFlushFreq, wait.NeverStop)
}

func FlushLogs() {
	klog.Flush()
}

func WithCRScheme(lg logr.Logger, obj metav1.Object, scheme *runtime.Scheme) logr.Logger {
	var gvk schema.GroupVersionKind

	if runtimeObj, ok := obj.(runtime.Object); ok {
		gvks, _, _ := scheme.ObjectKinds(runtimeObj)
		if len(gvks) > 0 {
			gvk = gvks[0]
		}
	}

	return lg.WithValues(
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
		"kind", gvk.Kind,
		"Version", gvk.Version,
	)
}

var contextKey = &struct{}{}

func IContext(ctx context.Context, names ...string) logr.Logger {
	l, err := logr.FromContext(ctx)
	if err != nil {
		l = Log
	}

	for _, n := range names {
		l = l.WithName(n)
	}

	return l
}
