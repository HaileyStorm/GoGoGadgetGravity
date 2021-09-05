package qt

import (
	"image"
	"image/png"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/therecipe/qt/gui"

	"GoGoGadgetGravity/physics"
)

// DrawParticles implements guis.GUIEnabler.DrawParticles. Unsurprisingly, it draws the provided particles in their
// current positions, and if enabled draws their position history trails.
func (q *Qt) DrawParticles(particles []*physics.Particle) {
	//timeStart := time.Now()

	q.StartIm2Qim(true)
	q.DrawViewBox()

	for _, p := range particles {
		// If TrackHistory is enabled, each historical position is drawn, with successively older positions
		// fainter (lower alpha)
		if p.TrackHistory() {
			for i, h := range p.PositionHistory() {
				q.drawFilledCircle(
					int(math.Round(h[0])),
					int(math.Round(h[1])),
					// Historical positions are drawn smaller
					int(math.Max(float64(p.Radius)*0.75, 1)),
					p.R, p.G, 0,
					// Calculate the alpha, which will have a minimum of 16 and a maximum
					// 16+240*((index-1)/HistorySize) - e.g. 232 if HistorySize is 10
					16+uint8((float64(p.A)-16)*(float64(i)/
						math.Min(float64(p.HistorySize()), float64(len(p.PositionHistory()))))))
			}
		}
		q.drawFilledCircle(int(math.Round(p.Position()[0])), int(math.Round(p.Position()[1])), p.Radius, p.R, p.G, 0, p.A)
	}
	// If not showing a (temporary) particle merge message, display the number of particles in the tatusbar
	if !strings.HasPrefix(q.statusbar.CurrentMessage(), "merging") {
		q.statusbar.ShowMessage("# of Particles: "+strconv.Itoa(len(particles)), 0)
	}

	//Threaded solution is slower in this situation...
	//Make each thread handle at least 10 particles so we're not over-threading
	/*maxThreads := math.Ceil(float64(len(particles)) / 10.0)
	threads := int(math.Min(10, maxThreads)) //10 threads, limited to maxThreads if smaller
	particlesPerThread := int(math.Floor(float64(len(particles)) / float64(threads)))
	var start, end int
	var wg sync.WaitGroup
	for i := 0; i < threads; i++ {
		start = i * particlesPerThread
		//For #particles not evenly divisible by #threads, make the last slice go all the way to the end
		if i == threads - 1 {
			end = len(particles)
		} else {
			end = start + particlesPerThread
		}
		wg.Add(1)
		go func(parts []*physics.Particle, wg *sync.WaitGroup) {
			defer wg.Done()
			for _, p := range parts {
				drawCircle(p.X, p.Y, p.Radius, p.R, p.G, 0, p.A)
			}
		}(particles[start:end], &wg)
	}
	wg.Wait()*/

	q.StopIm2Qim()

	//fmt.Println("DrawParticles time: " + time.Since(timeStart).String())
}

// DrawViewBox draws a box indicated the bounds/walls of the environment
func (q *Qt) DrawViewBox() {
	if !q.im2qim {
		q.Canvas = q.Pixmap.Pixmap().ToImage()
	}

	// Sides
	for _, x := range [2]int{0, q.EnvironmentSize - 1} {
		for y := 0; y < q.EnvironmentSize; y++ {
			q.setPixel(x, y, 0, 0, 255, 255)
		}
	}
	// Top & Bottom
	for _, y := range [2]int{0, q.EnvironmentSize - 1} {
		for x := 0; x < q.EnvironmentSize; x++ {
			q.setPixel(x, y, 0, 0, 255, 255)
		}
	}

	if !q.im2qim {
		q.Pixmap.SetPixmap(gui.NewQPixmap().FromImage(q.Canvas, 0))
	}
}

// drawCircleBorder draws a rasterized circle border (ring 1 pixel wide), centered on (cx, cy) and of the
// color provided by r,g,b,a, using the Midpoint Circle algorithm.
func (q *Qt) drawCircleBorder(cx, cy, rad int, r, g, b, a uint8) {
	// If circle falls entirely outside the environment, return
	if (cx+rad < 0 || cx-rad > q.EnvironmentSize) && (cy+rad < 0 || cy-rad > q.EnvironmentSize) {
		return
	}

	dx, dy, ex, ey := rad-1, 0, 1, 1
	err := ex - (rad * 2)

	for dx > dy {
		q.setPixel(cx+dx, cy+dy, r, g, b, a)
		q.setPixel(cx+dy, cy+dx, r, g, b, a)
		q.setPixel(cx-dy, cy+dx, r, g, b, a)
		q.setPixel(cx-dx, cy+dy, r, g, b, a)
		q.setPixel(cx-dx, cy-dy, r, g, b, a)
		q.setPixel(cx-dy, cy-dx, r, g, b, a)
		q.setPixel(cx+dy, cy-dx, r, g, b, a)
		q.setPixel(cx+dx, cy-dy, r, g, b, a)

		if err <= 0 {
			dy++
			err += ey
			ey += 2
		}
		if err > 0 {
			dx--
			ex += 2
			err += ex - (rad * 2)
		}
	}
}

// drawFilledCircle draws a filled-in (rasterized) circle, centered on (cx, cy) and of the color provided by r,g,b,a,
// using a (heavy) modification to the Midpoint Circle algorithm.
// This method is adapted from https://stackoverflow.com/q/10878209/5061881.
func (q *Qt) drawFilledCircle(cx, cy, rad int, r, g, b, a uint8) {
	// If circle falls entirely outside the environment, return
	if (cx+rad < 0 || cx-rad > q.EnvironmentSize) && (cy+rad < 0 || cy-rad > q.EnvironmentSize) {
		return
	}

	err, x, y := -rad, rad, 0
	var lastY int

	for x >= y {
		lastY = y
		err += y
		y++
		err += y

		q.drawTwoCenteredLines(cx, cy, x, lastY, r, g, b, a)

		if err >= 0 {
			if x != lastY {
				q.drawTwoCenteredLines(cx, cy, lastY, x, r, g, b, a)
			}

			err -= x
			x--
			err -= x
		}
	}
}

// drawTwoCenteredLines draws two lines of length 2*dx+1, centered on (cx,cy) and of the color provided by r,g,b,a,
// and with a gap of 2*dx-1 rows/pixels between them (that is, the line at cy and dy-1 lines to either side of it are
// not drawn).
// This is used by drawFilledCircle. See attribution there.
func (q *Qt) drawTwoCenteredLines(cx, cy, dx, dy int, r, g, b, a uint8) {
	q.drawHLine(cx-dx, cy+dy, cx+dx, r, g, b, a)
	if dy != 0 {
		q.drawHLine(cx-dx, cy-dy, cx+dx, r, g, b, a)
	}
}

// drawHLine draws a horizontal line from (x0,y0) to (x1,y0), of the color provided by r,g,b,a.
func (q *Qt) drawHLine(x0, y0, x1 int, r, g, b, a uint8) {
	for x := x0; x <= x1; x++ {
		q.setPixel(x, y0, r, g, b, a)
	}
}

// drawVLine draws a vertical line from (x0,y0) to (x0,y1), of the color provided by r,g,b,a.
func (q *Qt) drawVLine(x0, y0, y1 int, r, g, b, a uint8) {
	for y := y0; y <= y1; y++ {
		q.setPixel(x0, y, r, g, b, a)
	}
}

// setPixel sets the color of a single pixel
func (q *Qt) setPixel(x, y int, r, g, b, a uint8) {
	if q.im2qim {
		// Setting the pixel color bytes in the back-buffer is >5x the speed of img.Set()
		s := q.tempImage.PixOffset(x, y)
		if s < 0 || s >= len(q.tempImage.Pix) {
			return
		}

		// Locks are only necessary if multithreading (and not then if very rare write failures are acceptable - it's just a slice)
		//q.imgLock.Lock()
		q.tempImage.Pix[s], q.tempImage.Pix[s+1], q.tempImage.Pix[s+2], q.tempImage.Pix[s+3] = r, g, b, a
		//q.imgLock.Unlock()
	} else {
		q.Canvas.SetPixelColor2(x, y, gui.NewQColor3(int(r), int(g), int(b), int(a)))
	}
}

// StartIm2Qim enables im2qim mode for drawing on the Canvas (Canvas -> file -> standard library image)
func (q *Qt) StartIm2Qim(blank bool) {
	if blank {
		q.tempImage = image.NewNRGBA(image.Rect(0, 0, q.EnvironmentSize, q.EnvironmentSize))
	} else {
		// Write Canvas (a QImage) out to file and read it back to tempImage (an image.Image). Because I can't
		// figure out how to convert between the two using byte arrays etc.
		q.Canvas.Save("./tmp.png", "PNG", 100)
		reader, _ := os.Open("./tmp.png")
		p, _, _ := image.Decode(reader)
		q.tempImage, _ = p.(*image.NRGBA)
		reader.Close()
		os.Remove("./tmp.png")
	}

	q.im2qim = true
}

// StopIm2Qim disables im2qim mode for drawing on the Canvas (standard library image -> file -> canvas)
func (q *Qt) StopIm2Qim() {
	q.im2qim = false

	// Write tempImage (an image.Image) out to a file and read it back to Canvas (a QImage).
	out, _ := os.Create("./tmp.png")
	png.Encode(out, q.tempImage)
	q.Canvas.Load("./tmp.png", "")
	out.Close()
	os.Remove("./tmp.png")

	// For future efforts, something like the below seems like it should be close, but it doesn't work.
	/*var buf bytes.Buffer
	w := io.Writer(&buf)
	png.Encode(w, tempImage)
	Canvas.LoadFromData(buf.Bytes(), buf.Len(), "")*/

	q.Pixmap.SetPixmap(gui.NewQPixmap().FromImage(q.Canvas, 0))
}
