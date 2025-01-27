package detector

import (
	"context"
	"image"
	"image/color"

	"github.com/pkg/errors"
	"gocv.io/x/gocv"
)

var (
	rectColor   = color.RGBA{R: 0, G: 255, B: 0, A: 0}
	textColor   = color.RGBA{R: 0, G: 0, B: 255, A: 0}
	statusPoint = image.Pt(10, 20)
)

// Detector detects objects reading from a video device.
type Detector struct {
	video *gocv.VideoCapture

	firstFrame gocv.Mat
	frame      gocv.Mat
	gray       gocv.Mat
	delta      gocv.Mat
	thresh     gocv.Mat
	kernel     gocv.Mat

	handler HandleMotion

	streamer Streamer

	area float64
}

// Streamer holds stream methods for each type of image.
type Streamer interface {
	StreamDelta(img gocv.Mat)
	StreamFrame(img gocv.Mat)
	StreamThresh(img gocv.Mat)
}

// HandleMotion is the function that gets called when motion
// is detected.
type HandleMotion func(rect image.Rectangle)

// New creates a new detector, it opens the device specified by `deviceID`.
// The minimum size of the area in motion will be specified by `area`.
// Each type of image will be streamed to the streamer.
// `handler` will be called when motion is detected.
func New(deviceID int, area float64, handler HandleMotion, streamer Streamer) (*Detector, error) {
	video, err := gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		return nil, errors.Wrap(err, "Could not open capture device")
	}

	frame := gocv.NewMat()
	firstFrame := gocv.NewMat()
	if !video.Read(&frame) {
		return nil, errors.Wrap(err, "Could not read first video frame")
	}
	convertFrame(frame, &firstFrame)
	gocv.Flip(firstFrame, &firstFrame, 1)

	return &Detector{
		video:      video,
		firstFrame: firstFrame,
		frame:      frame,
		gray:       gocv.NewMat(),
		delta:      gocv.NewMat(),
		thresh:     gocv.NewMat(),
		kernel:     gocv.NewMat(),
		streamer:   streamer,
		handler:    handler,
		area:       area,
	}, nil
}

// Run runs the detector until the context is closed.
func (d *Detector) Run(ctx context.Context) {
	defer d.close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if d.scan() {
				return
			}
		}
	}
}

// scan scans the video for a new frame. It then parses this
// frame applying a few filters, thresholds and dilations in
// order to then calculate the contour of the area in movement.
// Once it has the contour it will draw the bounding rectangle
// and call the handle motion function.
func (d *Detector) scan() bool {
	if !d.video.Read(&d.frame) {
		return true
	}
	gocv.Flip(d.frame, &d.frame, 1)
	convertFrame(d.frame, &d.gray)

	gocv.AbsDiff(d.firstFrame, d.gray, &d.delta)
	gocv.Threshold(d.delta, &d.thresh, 50, 255, gocv.ThresholdBinary)
	gocv.Dilate(d.thresh, &d.thresh, d.kernel)
	cnt := bestContour(d.thresh.Clone(), d.area)
	if !cnt.IsNil() {
		rect := gocv.BoundingRect(cnt)
		gocv.Rectangle(&d.frame, rect, rectColor, 2)
		gocv.PutText(&d.frame, "Motion detected", statusPoint, gocv.FontHersheyPlain, 1.2, textColor, 2)
		d.handler(rect)
	}

	d.streamer.StreamFrame(d.frame)
	d.streamer.StreamDelta(d.delta)
	d.streamer.StreamThresh(d.thresh)

	return false
}

// close closes the detector.
func (d *Detector) close() error {
	d.firstFrame.Close()
	d.frame.Close()
	d.gray.Close()
	d.delta.Close()
	d.thresh.Close()
	d.kernel.Close()
	return d.video.Close()
}

// bestContour obtains the biggest contour in the frame(provided is bigger)
// than the minArea.
func bestContour(frame gocv.Mat, minArea float64) gocv.PointVector {
	cnts := gocv.FindContours(frame, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	var (
		bestCnt  gocv.PointVector
		bestArea = minArea
	)
	for i := 0; i < cnts.Size(); i++ {
		area := gocv.ContourArea(cnts.At(i))
		if area > bestArea {
			bestArea = area
			bestCnt = cnts.At(i)
		}
	}
	return bestCnt
}

func convertFrame(src gocv.Mat, dst *gocv.Mat) {
	gocv.Resize(src, &src, image.Point{X: 500, Y: 500}, 0, 0, gocv.InterpolationLinear)
	gocv.CvtColor(src, dst, gocv.ColorBGRToGray)
	gocv.GaussianBlur(*dst, dst, image.Point{X: 21, Y: 21}, 0, 0, gocv.BorderReflect101)
}
