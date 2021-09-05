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
				q.drawCircle(
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
		q.drawCircle(int(math.Round(p.Position()[0])), int(math.Round(p.Position()[1])), p.Radius, p.R, p.G, 0, p.A)
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

// drawCircle draws a filled-in circle
func (q *Qt) drawCircle(x, y, radius int, r, g, b, a uint8) {
	// If circle falls entirely outside the environment, return
	if (x+radius < 0 || x-radius > q.EnvironmentSize) && (y+radius < 0 || y-radius > q.EnvironmentSize) {
		return
	}

	if !q.im2qim {
		q.Canvas = q.Pixmap.Pixmap().ToImage()
	}

	var tx, ty int
	for i := 0; i <= radius; i++ {
		for j := 0; j <= radius; j++ {
			// This is more verbose than simply starting i&j at -1*radius and only doing the +/+ case,
			// but it's faster to do one radius check than four
			if i*i+j*j < radius*radius {
				tx = x + i
				ty = y + j
				q.setPixel(tx, ty, r, g, b, a)
				tx = x + i
				ty = y - j
				q.setPixel(tx, ty, r, g, b, a)
				tx = x - i
				ty = y - j
				q.setPixel(tx, ty, r, g, b, a)
				tx = x - i
				ty = y + j
				if tx >= 0 && ty >= 0 && tx < q.EnvironmentSize && ty < q.EnvironmentSize {
					q.setPixel(tx, ty, r, g, b, a)
				}
			}
		}
	}

	if !q.im2qim {
		q.Pixmap.SetPixmap(gui.NewQPixmap().FromImage(q.Canvas, 0))
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
