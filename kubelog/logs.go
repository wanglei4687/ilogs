package kubelog

import (
	"flag"
	"log"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var logFlushFreq = flag.Duration("log-flush-frequency", 5*time.Second, "Maximum number of seconds between log flushes")

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

	go wait.Until(klog.Flush, *logFlushFreq, wait.NerverStop)
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
