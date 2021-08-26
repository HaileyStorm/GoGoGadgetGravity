package physics

import (
	"math"
	"sort"

	"github.com/atedja/go-vector"
)

// InitializeParticles initializes all particles. It is used during restoration of state from file.
func InitializeParticles() {
	for _, p := range Engine.Particles {
		p.initialize()
	}
}

// SaveInitialParticleStates saves a copy of all particles in their current (initial generated / just restored
// from file) state, so they may be reverted to that state by the user during simulation.
func SaveInitialParticleStates() {
	Engine.initialParticles = make([]*Particle, len(Engine.Particles), len(Engine.Particles))
	for i, p := range Engine.Particles {
		Engine.initialParticles[i] = p.Clone()
	}
}

// RestoreInitialParticleStates restores all particles to the states stored in Engine.initialParticles by
// SaveInitialParticleStates, so the user may revert particles to their generated / restored from file states.
func RestoreInitialParticleStates() {
	Engine.Particles = make([]*Particle, len(Engine.initialParticles), len(Engine.initialParticles))
	for i, p := range Engine.initialParticles {
		Engine.Particles[i] = p.Clone()
	}
}

// UpdateParticles updates the Engine.Particles based on interactions between them (and the environment).
// Returns bools for whether a particle merge occurred (from a collision), whether >2 particles were involved,
// and the (largest) original particle & resulting merged particle.
func UpdateParticles() (bool, bool, *Particle, *Particle) {
	mergeOccurred, mergeMultiple := false, false
	var mergeSource, mergedResult *Particle

	updateParticleVelocities()
	updateParticlePositions()

	// Sort by mass. Used to merge to larger mass, and also a good order for drawing them.
	sort.Slice(Engine.Particles, func(i, j int) bool {
		return Engine.Particles[i].Mass() > Engine.Particles[j].Mass()
	})

	//region Handle Mergers
	if Engine.AllowMerge {
		// Particles to be deleted (in a given merger, all input particles are delete and newly merged particle added)
		// We track indexes instead of Particles in order to use the efficient removal method seen below
		var deleteList []int
		// Particles to be added
		var addList []*Particle
		var mergedParticle *Particle
		var mass, closeCharge, farCharge float64
		var position, velocity, tv vector.Vector
		var count float64

		for i, p := range Engine.Particles {
			if p.merging {
				count = float64(len(p.MergingWith))
				if count > 0 {
					mergeOccurred = true
					mergeSource = p

					if count > 1 {
						mergeMultiple = true
					}
					deleteList = append(deleteList, i)
					mass = p.Mass()
					// We average charges, weighted by mass of each particle
					closeCharge = p.CloseCharge() * mass
					farCharge = p.FarCharge() * mass
					// The position is also average & weighted, which we do by scaling each position vector by the
					// particle's mass, summing them, and then scaling the result back down by the total mass
					tv = p.Position().Clone()
					tv.Scale(mass)
					position = tv
					velocity = p.Velocity()
					//fmt.Printf("Merge. Original mass: %f, closeCharge: %f, farCharge: %f, position: %v,
					//velocity: %v\n", p.Mass(), p.CloseCharge(), p.FarCharge(), p.Position, p.Velocity)
					// Sum up the masses & charges
					for o, _ := range p.MergingWith {
						mass += o.Mass()
						closeCharge += o.CloseCharge() * o.Mass()
						farCharge += o.FarCharge() * o.Mass()
						tv = o.Position().Clone()
						tv.Scale(o.Mass())
						position = vector.Add(position, tv)
						tv = o.Velocity().Clone()
						tv.Scale(o.Mass() / p.Mass())
						velocity = vector.Add(velocity, tv)
						// We've merged from o to p, so we won't need to do p to o once we get to o (and indeed,
						// o will later be deleted)
						delete(o.MergingWith, p)
						//fmt.Printf("\tAdding particle w/: mass: %f, closeCharge: %f, farCharge: %f,
						//position: %v, velocity: %v\n", o.Mass(), o.CloseCharge(), o.FarCharge(), o.Position
						//o.Velocity)
					}

					// Compute the averages and create the new merged particle
					position.Scale(1.0 / mass)
					mergedParticle = NewParticle(mass, closeCharge/mass, farCharge/mass, position[0], position[1])
					mergedParticle.SetVelocity(velocity)
					// History data comes from the first (largest) particle involved in the merger
					mergedParticle.SetTrackHistory(p.TrackHistory())
					mergedParticle.SetHistorySize(p.HistorySize())
					mergedParticle.SetPositionHistory(p.PositionHistory())
					//fmt.Printf("Merge. New mass: %f, closeCharge: %f, farCharge: %f, position: %v, velocity: %v\n",
					//mergedParticle.Mass(), mergedParticle.CloseCharge(), mergedParticle.FarCharge(),
					//mergedParticle.Position, mergedParticle.Velocity)
					addList = append(addList, mergedParticle)
					// Returned for GUI display purposes
					mergedResult = mergedParticle
					// If the merge list for this particle has already been cleared by handling mergers from other
					// particles, then we just need to delete it
				} else {
					p.merging = false
					deleteList = append(deleteList, i)
				}
			}
		}

		// Delete original particles which have been merged into a new particle. Sort deleteList in decreasing order
		// so we can "move" each to be deleted item to the end of the slice and then truncate it
		sort.Slice(deleteList, func(i, j int) bool {
			return deleteList[i] > deleteList[j]
		})
		for _, i := range deleteList {
			// Move the last item to the ith position (so we keep it & overwrite the one we don't need)
			Engine.Particles[i] = Engine.Particles[len(Engine.Particles)-1]
			// Remove the last item (which we no longer need, as we have a copy of it in the ith position now)
			Engine.Particles = Engine.Particles[:len(Engine.Particles)-1]
		}

		// Add the newly created merged particles (the variadic call appends each item in addList separately - that is,
		// it doesn't try to append addList as a single new item)
		Engine.Particles = append(Engine.Particles, addList...)
	}
	//endregion Handle Mergers

	//region Wall bounce
	if Engine.WallBounce {
		var n vector.Vector
		var scale float64
		var err error
		var bounce bool
		for _, p := range Engine.Particles {
			bounce = false
			// If the circle representing the particle extends beyond the sides...
			if int(p.Position()[0])-p.Radius < 0 || int(p.Position()[0])+p.Radius > Engine.EnvironmentSize-1 {
				// p.Velocity - n, where n is scaled by 2* the dot product of p.Velocity & n, reflects p.Velocity over
				// (n rotated by 90 degrees). So n is horizontal, so that the reflection happens over a vertical line.
				n = vector.NewWithValues([]float64{1, 0})
				scale, err = vector.Dot(p.Velocity(), n)
				if err == nil {
					// Make sure the particle didn't go past the edge
					p.Position()[0] = math.Max(float64(p.Radius), math.Min(p.Position()[0],
						float64(Engine.EnvironmentSize)-float64(p.Radius)-1))
					bounce = true
				}
			}
			// If not already bouncing on sides and the circle representing the particle extends beyond the
			// top or bottom...
			if !bounce && (int(p.Position()[1])-p.Radius < 0 ||
				int(p.Position()[1])+p.Radius > Engine.EnvironmentSize-1) {
				// p.Velocity - n, where n is scaled by 2* the dot product of p.Velocity & n, reflects p.Velocity over
				// (n rotated by 90 degrees). So n is vertical, so that the reflection happens over a horizontal line.
				n = vector.NewWithValues([]float64{0, 1})
				scale, err = vector.Dot(p.Velocity(), n)
				if err == nil {
					// Make sure the particle didn't go past the edge
					p.Position()[1] = math.Max(float64(p.Radius), math.Min(p.Position()[1],
						float64(Engine.EnvironmentSize)-float64(p.Radius)-1))
					bounce = true
				}
			}
			// Complete the reflection
			if bounce {
				scale *= 2
				n.Scale(scale)
				p.SetVelocity(vector.Subtract(p.Velocity(), n))
			}
		}
	}
	//endregion Wall bounce

	return mergeOccurred, mergeMultiple, mergeSource, mergedResult
}

// updateParticleVelocities updates the Engine.Particles velocities by calculating and summing the three force
// acceleration vectors acting on the Particle (based on the relative positions, masses, and charges of all other
// Particles) and adding that to the current Particle's current Velocity.
func updateParticleVelocities() {
	var v, vc, vf, g, c, f vector.Vector
	var mag float64

	for _, p := range Engine.Particles {
		// Force acceleration vectors (average of force vectors between p and each other particle it isn't merging with
		// or bouncing against)
		g = vector.New(2)
		c = vector.New(2)
		f = vector.New(2)
		// Count of particles for which force interactions with p are calculated (for averaging)
		ct := 0

		// Work with p against every other particle (o)
		for _, o := range Engine.Particles {
			// If comparing against itself, or p & o are merging, we don't need to calculate their force effects
			// on each other
			if _, ok := p.MergingWith[o]; ok || p == o {
				continue
			}

			// Get the distance (mag) between the two particles
			v = vector.Subtract(p.Position(), o.Position())
			mag = v.Magnitude()

			// Stop bounce once separated
			if p.bouncing && p.bouncingAgainst == o {
				if mag > Engine.bounceCompleteDistFactor*float64(p.Radius+o.Radius) {
					p.bouncing = false
				}
				continue
			}

			// New collision (not already bouncing against each other and distance between them is less than
			// combined radii) - determine if merge or bounce
			if !(p.bouncing && p.bouncingAgainst == o) && mag < float64(p.Radius+o.Radius) {
				var massRatio float64
				if Engine.AllowMerge {
					if p.Mass() > o.Mass() {
						massRatio = p.Mass() / o.Mass()
					} else {
						massRatio = o.Mass() / p.Mass()
					}
				}

				// Merge if mergers are enabled and the mass difference is sufficient and the close charge doesn't repel
				// enough to prevent it
				if Engine.AllowMerge && massRatio > Engine.mergeMassRatioThreshold &&
					(math.Signbit(p.CloseCharge()) != math.Signbit(o.CloseCharge()) ||
						math.Abs(p.CloseCharge())+math.Abs(o.CloseCharge()) < Engine.mergeCloseChargeThreshold) {
					p.merging = true
					// Add o to p's MergingWith (set its value to an empty anonymous struct, so that the key exists)
					p.MergingWith[o] = struct{}{}
					// If o doesn't already have p in it's MergingWith (because o came before p in the outer loop),
					// add it
					if _, ok := o.MergingWith[p]; !ok {
						o.merging = true
						o.MergingWith[p] = struct{}{}
					}
					// Bounce (see WallBounce logic in UpdateParticles for vector math description, except the direction of
					// the reflecting vector is determined by which axis the particle's are moving along most, rather than
					// which wall they're bouncing against)
				} else {
					var n vector.Vector
					// Todo: this isn't quite right. I think perhaps we need to account for whether the (primary axis)
					// velocities of the two particles are in the same or opposite directions ... and then multiply
					// the reflection vector by -1 if ... same??
					if math.Abs(p.Velocity()[0])+math.Abs(o.Velocity()[0]) >
						math.Abs(p.Velocity()[1])+math.Abs(o.Velocity()[1]) {
						n = vector.NewWithValues([]float64{0, 1})
					} else {
						n = vector.NewWithValues([]float64{1, 0})
					}

					scale, err := vector.Dot(p.Velocity(), n)
					if err != nil {
						continue
					}
					// We now know the math of the bounce will succeed, so it's safe to set the bouncing state
					// (which gets unset when the particles are sufficiently separated)
					p.bouncing = true
					p.bouncingAgainst = o
					scale *= 2
					n.Scale(scale)
					p.SetVelocity(vector.Subtract(p.Velocity(), n))
				}
				// If we have a new collision (bounce/merge), we don't need to calculate the forces between p & o
				// (which happens below)
				continue
			}
			// Increment the total number of particles for which forces are calculated between p & said particles,
			// so that the forces can be averaged
			ct++

			// v is the vector between p & o, which we need for calculating force vectors between the two.
			// We need to a copy of it for each force (v for gravity, vc for close charge, vf for far charge)
			vc = v.Clone()
			vf = v.Clone()

			// Simplified formula for getting v's unit vector (v/mag) and then scaling it by the
			// felt force acceleration: f=G*m1*m2/mag^2 and a=f/m (own particle's mass divides out)
			v.Scale((Engine.GravityStrength * o.Mass() * -1) / math.Pow(mag, 3))
			g = vector.Add(g, v)

			// Simplified formula for getting vc's unit vector (vc/mag) and then scaling it by the
			// felt force acceleration: f=C*c1*c2/mag^3 and a=f/m
			vc.Scale((Engine.CloseChargeStrength * p.CloseCharge() * o.CloseCharge()) /
				(p.Mass() * math.Pow(mag, 4)))
			c = vector.Add(c, vc)

			// Simplified formula for getting vf's unit vector (vf/mag) and then scaling it by the
			// felt force acceleration: f=C*c1*c2*mag and a=f/m (the distance divides out since proportional to
			// distance rather than inversely and scaling to unit vector puts the magnitude on the divisor).
			vf.Scale((Engine.FarChargeStrength * p.FarCharge() * o.FarCharge() * -1) / p.Mass())
			f = vector.Add(f, vf)
		}

		// Compute the average force acceleration vectors
		g.Scale(1.0 / float64(ct))
		c.Scale(1.0 / float64(ct))
		f.Scale(1.0 / float64(ct))

		// Sum the (now averaged) acceleration vectors from each force and apply it to the particle
		// (add the summed acceleration vector to the velocity)
		p.SetVelocity(vector.Add(vector.Add(vector.Add(p.Velocity(), g), c), f))
	}
}

// updateParticlePositions updates the Engine.Particles positions by calling Particle.UpdatePosition on each particle
// (which adds the Particle's Velocity vector to its Position vector).
func updateParticlePositions() {
	for _, p := range Engine.Particles {
		p.UpdatePosition()
	}
}
