package app

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type promtoolBackfillService struct {
	dataDirectory string
	promtoolPath  string
	port          int
	log           logrus.FieldLogger
}

const uploadTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <title>Upload Openmetrics File</title>
  </head>
  <body>
    <form
      enctype="multipart/form-data"
      action="upload"
      method="post"
    >
      <input type="file" name="metrics" />
      <input type="submit" value="upload" />
   </form>
  </body>
</html>
`

func NewCommand(log logrus.FieldLogger) *cobra.Command {
	app := &promtoolBackfillService{log: log}
	cmd := &cobra.Command{
		Use:   "promtool-backfill-service",
		Short: "Prometheus promtool backfill service",
		RunE:  app.run,
	}
	cmd.Flags().IntVar(&app.port, "port", 9100, "http port")
	cmd.Flags().StringVar(&app.dataDirectory, "data-directory", "data", "prometheus data directory")
	cmd.Flags().StringVar(&app.promtoolPath, "promtool-path", "promtool", "file path to promtool")
	return cmd
}

func (svc *promtoolBackfillService) run(cmd *cobra.Command, args []string) error {
	http.HandleFunc("/upload", svc.uploadHandler)
	svc.log.Infof("listening on port %d", svc.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", svc.port), nil)
}

func (svc *promtoolBackfillService) uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		svc.displayUpload(w)
	case "POST":
		svc.uploadFile(w, r)
	}
}

func (svc *promtoolBackfillService) displayUpload(w http.ResponseWriter) {
	_, _ = w.Write([]byte(uploadTemplate))
}

func (svc *promtoolBackfillService) uploadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 * 1024 * 1024) // store file up to 10 MB in memory
	if err != nil {
		svc.error(w, "parse failed", err)
		return
	}

	file, handler, err := r.FormFile("metrics")
	if err != nil {
		svc.error(w, "metrics file not found in form", err)
		return
	}
	defer file.Close()

	dst, err := ioutil.TempFile("", handler.Filename)
	if err != nil {
		svc.error(w, "tempfile naming failed", err)
		return
	}
	defer dst.Close()
	defer os.Remove(dst.Name())

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		svc.error(w, "copying file failed", err)
		return
	}

	if err := dst.Close(); err != nil {
		svc.error(w, "closing file failed", err)
		return
	}

	_, fname := path.Split(dst.Name())
	svc.log.Infof("Uploaded file name=%s, size=%d (%s)", handler.Filename, handler.Size, fname)

	if err := svc.backfill(dst); err != nil {
		svc.error(w, "backfilling file failed", err)
		return
	}

	w.Write([]byte(fmt.Sprintf("backfilled %s", handler.Filename)))
	svc.log.Infof("backfill successful for name=%s, size=%d (%s)", handler.Filename, handler.Size, fname)
}

func (svc *promtoolBackfillService) backfill(file *os.File) error {
	var stderr bytes.Buffer
	cmd := exec.Command(svc.promtoolPath, "tsdb", "create-blocks-from", "openmetrics", file.Name(), svc.dataDirectory)
	cmd.Stderr = &stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("promtool tsdb create-blocks-from execution failed: %s\nDetails: %s", err, stderr.String())
	}
	return nil
}

func (svc *promtoolBackfillService) error(w http.ResponseWriter, msg string, err error) {
	svc.log.Errorf("%s: %s", msg, err)
	http.Error(w, msg, http.StatusInternalServerError)
}
