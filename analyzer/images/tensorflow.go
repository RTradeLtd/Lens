package images

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
)

// TensorflowAnalyzer represents a wrapper around a Tensorflow-based analyzer
type TensorflowAnalyzer interface {
	Analyze(jobID string, content []byte) (category string, err error)
}

// All credits for this go to the developers of the example in the following link
// https://godoc.org/github.com/tensorflow/tensorflow/tensorflow/go
// This is simply a modified version, intended to be run as a analyzer method by the Lens service

// Analyzer is used to analyze images
type Analyzer struct {
	session    *tf.Session
	graph      *tf.Graph
	labelsFile string

	l *zap.SugaredLogger
}

// ConfigOpts is used to configure our image analyzer
type ConfigOpts struct {
	ModelLocation string `json:"model_location"`
}

// NewAnalyzer is used to analyze an image and classify it
func NewAnalyzer(opts ConfigOpts, logger *zap.SugaredLogger) (*Analyzer, error) {
	// load a seralized graph definition
	modelFile, labelsFile, err := modelFiles(opts.ModelLocation)
	if err != nil {
		return nil, err
	}
	model, err := ioutil.ReadFile(modelFile)
	if err != nil {
		return nil, err
	}
	// create the graph in memory
	graph := tf.NewGraph()
	if err = graph.Import(model, ""); err != nil {
		return nil, err
	}
	// create a session
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	// defer session closure
	// defer session.Close()
	return &Analyzer{
		session:    session,
		labelsFile: labelsFile,
		graph:      graph,
	}, nil
}

// Analyze is used to run an image against the Inception v5 pre-trained model
func (a *Analyzer) Analyze(jobID string, content []byte) (string, error) {
	tensor, err := makeTensorFromImage(content)
	if err != nil {
		return "", err
	}
	output, err := a.session.Run(
		map[tf.Output]*tf.Tensor{
			a.graph.Operation("input").Output(0): tensor,
		},
		[]tf.Output{
			a.graph.Operation("output").Output(0),
		},
		nil,
	)
	if err != nil {
		return "", err
	}
	probabilities := output[0].Value().([][]float32)[0]
	return a.classify(probabilities, a.labelsFile)
}

func (a *Analyzer) classify(probabilities []float32, labelsFile string) (string, error) {
	bestIdx := 0
	for i, p := range probabilities {
		if p > probabilities[bestIdx] {
			bestIdx = i
		}
	}
	// Found the best match. Read the string from labelsFile, which
	// contains one line per label.
	file, err := os.Open(labelsFile)
	if err != nil {
		return "", nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var labels []string
	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("ERROR: failed to read %s: %v", labelsFile, err)
	}
	return labels[bestIdx], nil
}

// Convert the image in filename to a Tensor suitable as input to the Inception model.
func makeTensorFromImage(content []byte) (*tf.Tensor, error) {
	// DecodeJpeg uses a scalar String-valued tensor as input.
	tensor, err := tf.NewTensor(string(content))
	if err != nil {
		return nil, err
	}
	// Construct a graph to normalize the image
	graph, input, output, err := constructGraphToNormalizeImage()
	if err != nil {
		return nil, err
	}
	// Execute that graph to normalize this one image
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	normalized, err := session.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		return nil, err
	}
	return normalized[0], nil
}

// The inception model takes as input the image described by a Tensor in a very
// specific normalized format (a particular image size, shape of the input tensor,
// normalized pixel values etc.).
//
// This function constructs a graph of TensorFlow operations which takes as
// input a JPEG-encoded string and returns a tensor suitable as input to the
// inception model.
func constructGraphToNormalizeImage() (graph *tf.Graph, input, output tf.Output, err error) {
	// Some constants specific to the pre-trained model at:
	// https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip
	//
	// - The model was trained after with images scaled to 224x224 pixels.
	// - The colors, represented as R, G, B in 1-byte each were converted to
	//   float using (value - Mean)/Scale.
	const (
		H, W  = 224, 224
		Mean  = float32(117)
		Scale = float32(1)
	)
	// - input is a String-Tensor, where the string the JPEG-encoded image.
	// - The inception model takes a 4D tensor of shape
	//   [BatchSize, Height, Width, Colors=3], where each pixel is
	//   represented as a triplet of floats
	// - Apply normalization on each pixel and use ExpandDims to make
	//   this single image be a "batch" of size 1 for ResizeBilinear.
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)
	output = op.Div(s,
		op.Sub(s,
			op.ResizeBilinear(s,
				op.ExpandDims(s,
					op.Cast(s,
						op.DecodeJpeg(s, input, op.DecodeJpegChannels(3)), tf.Float),
					op.Const(s.SubScope("make_batch"), int32(0))),
				op.Const(s.SubScope("size"), []int32{H, W})),
			op.Const(s.SubScope("mean"), Mean)),
		op.Const(s.SubScope("scale"), Scale))
	graph, err = s.Finalize()
	return graph, input, output, err
}

func modelFiles(dir string) (modelfile, labelsfile string, err error) {
	const URL = "https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip"
	var (
		model   = filepath.Join(dir, "tensorflow_inception_graph.pb")
		labels  = filepath.Join(dir, "imagenet_comp_graph_label_strings.txt")
		zipfile = filepath.Join(dir, "inception5h.zip")
	)
	if filesExist(model, labels) == nil {
		return model, labels, nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", err
	}
	if err := download(URL, zipfile); err != nil {
		return "", "", fmt.Errorf("failed to download %v - %v", URL, err)
	}
	if err := unzip(dir, zipfile); err != nil {
		return "", "", fmt.Errorf("failed to extract contents from model archive: %v", err)
	}
	os.Remove(zipfile)
	return model, labels, filesExist(model, labels)
}

func filesExist(files ...string) error {
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			return fmt.Errorf("unable to stat %s: %v", f, err)
		}
	}
	return nil
}

func download(URL, filename string) error {
	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	return err
}

func unzip(dir, zipfile string) error {
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		src, err := f.Open()
		if err != nil {
			return err
		}
		dst, err := os.OpenFile(filepath.Join(dir, f.Name), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
		dst.Close()
	}
	return nil
}
