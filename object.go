//=============================================================
// object.go
//-------------------------------------------------------------
// Different objects that are subject to physics and flood fill
// for destuction into pieces. Much like mob, but special adds
// like FF and different physics and actions.
//=============================================================
package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"image"
)

type object struct {
	textureFile string
	img         image.Image
	model       *imdraw.IMDraw
	canvas      *pixelgl.Canvas
	bounds      *Bounds
	objectType  entityType
	mass        float64
	restitution float64
	height      int
	width       int
	size        int
	force       pixel.Vec
	pixels      []uint32
	prevPos     []pixel.Vec
	bounces     int
	vx          float64
	vy          float64
	fx          float64
	fy          float64
	scale       float64
}

//=============================================================
//
//=============================================================
func (o *object) create(x_, y_ float64) {
	o.prevPos = make([]pixel.Vec, 100)

	o.mass = 5
	o.restitution = -0.3
	o.fx = 1
	o.fy = 1
	o.vx = 1
	o.vy = 1

	o.img, o.width, o.height, o.size = loadTexture(o.textureFile)

	// Initiate bounds for qt
	o.bounds = &Bounds{
		X:      x_,
		Y:      y_,
		Width:  float64(o.width) * o.scale,
		Height: float64(o.height) * o.scale,
		entity: Entity(o),
	}

	o.pixels = make([]uint32, o.size*o.size)

	for x := 0; x < o.width; x++ {
		for y := 0; y < o.height; y++ {
			r, g, b, a := o.img.At(x, o.height-y).RGBA()
			o.pixels[x*o.size+y] = r&0xFF<<24 | g&0xFF<<16 | b&0xFF<<8 | a&0xFF
		}
	}

	// Generate some CD pixel for faster CD check.
	//rand.Seed(time.Now().UTC().UnixNano())
	//for x := 0; x < 20; x++ {
	//	o.cdPixels = append(o.cdPixels, [2]uint32{uint32(rand.Intn(o.width)), uint32(rand.Intn(o.height))})
	//}

	o.canvas = pixelgl.NewCanvas(pixel.R(0, 0, float64(o.width), float64(o.height)))

	// build initial
	o.build()

	// Add object to QT
	global.gWorld.AddObject(o.bounds)
}

//=============================================================
// Build
//=============================================================
func (o *object) build() {
	o.model = imdraw.New(nil)
	for x := 0; x < o.width; x++ {
		for y := 0; y < o.height; y++ {
			p := o.pixels[x*o.size+y]
			if p == 0 {
				continue
			}

			o.model.Color = pixel.RGB(
				float64(p>>24&0xFF)/255.0,
				float64(p>>16&0xFF)/255.0,
				float64(p>>8&0xFF)/255.0,
			).Mul(pixel.Alpha(float64(p&0xFF) / 255.0))
			o.model.Push(
				pixel.V(float64(x*wPixelSize), float64(y*wPixelSize)),
				pixel.V(float64(x*wPixelSize+wPixelSize), float64(y*wPixelSize+wPixelSize)),
			)
			o.model.Rectangle(0)
		}
	}

	o.canvas.Clear(pixel.RGBA{0, 0, 0, 0})
	o.model.Draw(o.canvas)
}

//=============================================================
//
//  Function to implement Entity interface
//
//=============================================================
//=============================================================
//
//=============================================================
func (o *object) hit(x_, y_ float64) bool {
	//x := int(math.Abs(float64(o.bounds.X - x_)))
	//y := int(math.Abs(float64(o.bounds.Y - y_)))

	o.build()
	return true
}

//=============================================================
//
//=============================================================
func (o *object) explode() {
}

//=============================================================
//
//=============================================================
func (o *object) move(dx, dy float64) {
	// Add the force, movenment is handled in the physics function
	// o.force.X += dx * o.speed
	// o.force.Y += dy * o.speed
}

//=============================================================
//
//=============================================================
func (o *object) getPosition() pixel.Vec {
	return pixel.Vec{o.bounds.X, o.bounds.Y}
}

//=============================================================
//
//=============================================================
func (o *object) getMass() float64 {
	return o.mass
}

//=============================================================
//
//=============================================================
func (o *object) getType() entityType {
	return entityObject
}

//=============================================================
//
//=============================================================
func (o *object) setPosition(x, y float64) {
	o.bounds.X = x
	o.bounds.Y = y
}

//=============================================================
//
//=============================================================
func (o *object) saveMove() {
	//o.prevPos = append(o.prevPos, pixel.Vec{o.bounds.X, o.bounds.Y})
	//// TBD: Only remove every second or something
	//if len(o.prevPos) > 100 {
	//	o.prevPos = o.prevPos[:100]
	//}
}

//=============================================================
// Physics
//=============================================================
func (o *object) physics(dt float64) {

	if global.gWorld.IsWall(o.bounds.X, o.bounds.Y) {
		o.bounces++
		if o.bounces <= 4 {
			if o.vy < 0 {
				o.vy *= o.restitution
			} else {
				if o.vx > 0 {
					o.vx *= -o.restitution
					o.vy *= -o.restitution
				} else if o.vx < 0 {
					o.vx *= -o.restitution
					o.vy *= -o.restitution
				}
			}
		} else {
			o.fx = 0
			o.fy = 0
		}
	}
	//o.saveMove()
	ax := o.fx * dt * o.vx * o.mass
	ay := o.fy * dt * o.vy * o.mass
	o.bounds.X += ax
	o.bounds.Y += ay

	o.vy -= dt * o.fy

	if o.fx > 0 {
		o.fx -= dt * global.gWorld.gravity * o.mass
	} else {
		o.fx = 0
	}
	if o.fy > 0 {
		o.fy -= dt * global.gWorld.gravity * o.mass
	}

}

//=============================================================
//
//=============================================================
func (o *object) draw(dt float64) {
	// Update physics
	o.physics(dt)

	o.canvas.Draw(global.gWin, pixel.IM.ScaledXY(pixel.ZV, pixel.V(o.scale, o.scale)).Moved(pixel.V(o.bounds.X+o.bounds.Width/2, o.bounds.Y+o.bounds.Height/2)))
	o.unStuck(dt)
}

//=============================================================
// Unstuck the objet if stuck.
//=============================================================
func (o *object) unStuck(dt float64) {
	bottom := false
	top := false
	offset := 1.0
	// Check bottom pixels
	for x := o.bounds.X; x < o.bounds.X+o.bounds.Width; x += 2 {
		if global.gWorld.IsRegular(x, o.bounds.Y+offset) {
			bottom = true
			break
		}
	}

	//Check top pixels
	for x := o.bounds.X; x < o.bounds.X+o.bounds.Width; x += 2 {
		if global.gWorld.IsRegular(x, o.bounds.Y+o.bounds.Height-offset) {
			top = true
			break
		}
	}

	if bottom {
		o.bounds.Y += 10 * o.mass * dt
	} else if top {
		o.bounds.Y -= 10 * o.mass * dt
	}
}