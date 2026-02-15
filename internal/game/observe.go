package game

import "math"

// --- Types for AI bridge ---

// AIAction represents the set of actions an AI agent can take each tick.
type AIAction struct {
	RotateLeft  bool
	RotateRight bool
	Thrust      bool
	Shoot       bool
	Hyperspace  bool
}

// AIAgent is the interface that an AI controller must implement.
type AIAgent interface {
	Act(observation []float64) AIAction
}

// --- World accessors (read-only, exported) ---

// PlayerPosition returns the player's X, Y position.
func (w *World) PlayerPosition() (float64, float64) {
	pos := w.positions[w.Player]
	if pos == nil {
		return 0, 0
	}
	return pos.X, pos.Y
}

// PlayerVelocity returns the player's velocity.
func (w *World) PlayerVelocity() (float64, float64) {
	vel := w.velocities[w.Player]
	if vel == nil {
		return 0, 0
	}
	return vel.X, vel.Y
}

// PlayerAngle returns the player's rotation angle.
func (w *World) PlayerAngle() float64 {
	rot := w.rotations[w.Player]
	if rot == nil {
		return 0
	}
	return rot.Angle
}

// PlayerInvulnerable returns whether the player is currently invulnerable.
func (w *World) PlayerInvulnerable() bool {
	pc := w.players[w.Player]
	if pc == nil {
		return false
	}
	return pc.Invulnerable
}

// AsteroidCount returns the number of active asteroids.
func (w *World) AsteroidCount() int {
	return len(w.asteroids)
}

// AsteroidInfo holds exported asteroid data for AI observation.
type AsteroidInfo struct {
	X, Y, VX, VY, Radius float64
	Size                  AsteroidSize
}

// Asteroids returns info about all active asteroids.
func (w *World) Asteroids() []AsteroidInfo {
	result := make([]AsteroidInfo, 0, len(w.asteroids))
	for e, at := range w.asteroids {
		pos := w.positions[e]
		vel := w.velocities[e]
		col := w.colliders[e]
		if pos == nil || col == nil {
			continue
		}
		vx, vy := 0.0, 0.0
		if vel != nil {
			vx, vy = vel.X, vel.Y
		}
		result = append(result, AsteroidInfo{
			X: pos.X, Y: pos.Y,
			VX: vx, VY: vy,
			Radius: col.Radius,
			Size:   at.Size,
		})
	}
	return result
}

// SaucerInfo holds exported saucer data for AI observation.
type SaucerInfo struct {
	X, Y, VX, VY, Radius float64
	Size                  SaucerSize
}

// ActiveSaucer returns info about the active saucer, or nil if none.
func (w *World) ActiveSaucer() *SaucerInfo {
	if w.SaucerActive == 0 || !w.Alive(w.SaucerActive) {
		return nil
	}
	e := w.SaucerActive
	pos := w.positions[e]
	vel := w.velocities[e]
	col := w.colliders[e]
	st := w.saucers[e]
	if pos == nil || col == nil || st == nil {
		return nil
	}
	vx, vy := 0.0, 0.0
	if vel != nil {
		vx, vy = vel.X, vel.Y
	}
	return &SaucerInfo{
		X: pos.X, Y: pos.Y,
		VX: vx, VY: vy,
		Radius: col.Radius,
		Size:   st.Size,
	}
}

// SaucerBulletInfo holds exported saucer bullet data.
type SaucerBulletInfo struct {
	X, Y, VX, VY float64
}

// SaucerBullets returns info about all active saucer bullets.
func (w *World) SaucerBullets() []SaucerBulletInfo {
	result := make([]SaucerBulletInfo, 0, len(w.saucerBullets))
	for e := range w.saucerBullets {
		pos := w.positions[e]
		vel := w.velocities[e]
		if pos == nil {
			continue
		}
		vx, vy := 0.0, 0.0
		if vel != nil {
			vx, vy = vel.X, vel.Y
		}
		result = append(result, SaucerBulletInfo{
			X: pos.X, Y: pos.Y,
			VX: vx, VY: vy,
		})
	}
	return result
}

// BulletInfo holds exported player bullet data.
type BulletInfo struct {
	X, Y float64
}

// PlayerBullets returns info about all active player bullets.
func (w *World) PlayerBullets() []BulletInfo {
	result := make([]BulletInfo, 0, len(w.bullets))
	for e := range w.bullets {
		pos := w.positions[e]
		if pos == nil {
			continue
		}
		result = append(result, BulletInfo{X: pos.X, Y: pos.Y})
	}
	return result
}

// --- AI Input System ---

// AIInputSystem applies an AIAction to the player, mirroring InputSystem but
// reading from the action struct instead of the keyboard.
func AIInputSystem(w *World, action AIAction) {
	for e, pc := range w.players {
		rot := w.rotations[e]
		vel := w.velocities[e]

		if action.RotateLeft {
			rot.Angle -= rotationSpeed
		}
		if action.RotateRight {
			rot.Angle += rotationSpeed
		}

		pc.Thrusting = action.Thrust
		if pc.Thrusting {
			vel.X += math.Cos(rot.Angle) * thrustPower
			vel.Y += math.Sin(rot.Angle) * thrustPower
			speed := math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)
			if speed > maxSpeed {
				vel.X = vel.X / speed * maxSpeed
				vel.Y = vel.Y / speed * maxSpeed
			}
		}

		vel.X *= friction
		vel.Y *= friction

		pc.ShootPressed = action.Shoot
		pc.HyperspacePressed = action.Hyperspace
	}
}

// --- Observation extraction ---

// WrapDelta computes the shortest signed distance on a toroidal axis of length size.
func WrapDelta(a, b, size float64) float64 {
	d := b - a
	if d > size/2 {
		d -= size
	} else if d < -size/2 {
		d += size
	}
	return d
}

// halfDiag is half the screen diagonal, used for normalizing distances.
var halfDiag = math.Sqrt(ScreenWidth*ScreenWidth+ScreenHeight*ScreenHeight) / 2

// ObservationSize is the length of the observation vector.
const ObservationSize = 40

// ExtractObservation builds the normalized observation vector from the current
// world state. This lives in the game package because it needs direct access
// to World internals and game constants.
//
// Layout (40 floats):
//
//	 0-1   Player X, Y               [-1, 1]
//	 2-3   Player VX, VY             / maxSpeed
//	 4-5   sin(angle), cos(angle)    [-1, 1]
//	 6     Invulnerable              0 or 1
//	 7     Bullet count              / MaxPlayerBullets
//	 8-27  5 nearest asteroids × 4   (rel dx, dy / halfDiag, dvx, dvy / maxSpeed)
//	28-30  Saucer present, rel dx, dy
//	31-32  Nearest saucer bullet rel dx, dy
//	33     Lives                     / 5
//	34     LEAD angle error to nearest asteroid / π  → [-1, 1]
//	35     Closing speed to nearest asteroid / maxSpeed
//	36     Lead line-of-fire quality [0, 1]
//	37     Nearest asteroid distance / halfDiag
//	38     Collision threat urgency [0, 1] (0=safe, 1=imminent impact)
//	39     Dodge angle error / π [-1, 1] (which way to thrust to escape)
func ExtractObservation(w *World) []float64 {
	obs := make([]float64, ObservationSize)

	// Player state
	px, py := w.PlayerPosition()
	obs[0] = (px/ScreenWidth)*2 - 1
	obs[1] = (py/ScreenHeight)*2 - 1

	vx, vy := w.PlayerVelocity()
	obs[2] = vx / maxSpeed
	obs[3] = vy / maxSpeed

	angle := w.PlayerAngle()
	obs[4] = math.Sin(angle)
	obs[5] = math.Cos(angle)

	if w.PlayerInvulnerable() {
		obs[6] = 1
	}

	obs[7] = float64(w.BulletCount()) / float64(MaxPlayerBullets)

	// Collect asteroid data with distances
	type asteroidObs struct {
		dx, dy         float64 // wrap-aware delta from player
		dvx, dvy       float64 // relative velocity (asteroid - player)
		avx, avy       float64 // absolute asteroid velocity
		dist, radius   float64
	}
	rawAsteroids := w.Asteroids()
	aobs := make([]asteroidObs, 0, len(rawAsteroids))
	for _, a := range rawAsteroids {
		dx := WrapDelta(px, a.X, ScreenWidth)
		dy := WrapDelta(py, a.Y, ScreenHeight)
		d := math.Sqrt(dx*dx + dy*dy)
		dvx := a.VX - vx
		dvy := a.VY - vy
		aobs = append(aobs, asteroidObs{dx, dy, dvx, dvy, a.VX, a.VY, d, a.Radius})
	}

	// Sort top-5 nearest
	for i := 0; i < min(5, len(aobs)); i++ {
		minIdx := i
		for j := i + 1; j < len(aobs); j++ {
			if aobs[j].dist < aobs[minIdx].dist {
				minIdx = j
			}
		}
		aobs[i], aobs[minIdx] = aobs[minIdx], aobs[i]
	}

	// 5 nearest asteroids (indices 8-27, 4 floats each)
	for i := 0; i < 5; i++ {
		base := 8 + i*4
		if i < len(aobs) {
			obs[base+0] = aobs[i].dx / halfDiag
			obs[base+1] = aobs[i].dy / halfDiag
			obs[base+2] = aobs[i].dvx / maxSpeed
			obs[base+3] = aobs[i].dvy / maxSpeed
		}
	}

	// Saucer (indices 28-30)
	saucer := w.ActiveSaucer()
	if saucer != nil {
		obs[28] = 1
		obs[29] = WrapDelta(px, saucer.X, ScreenWidth) / halfDiag
		obs[30] = WrapDelta(py, saucer.Y, ScreenHeight) / halfDiag
	}

	// Nearest saucer bullet (indices 31-32)
	sBullets := w.SaucerBullets()
	if len(sBullets) > 0 {
		bestDist := math.Inf(1)
		bestDx, bestDy := 0.0, 0.0
		for _, sb := range sBullets {
			dx := WrapDelta(px, sb.X, ScreenWidth)
			dy := WrapDelta(py, sb.Y, ScreenHeight)
			d := math.Sqrt(dx*dx + dy*dy)
			if d < bestDist {
				bestDist = d
				bestDx, bestDy = dx, dy
			}
		}
		obs[31] = bestDx / halfDiag
		obs[32] = bestDy / halfDiag
	}

	// Lives (index 33)
	obs[33] = float64(w.Lives) / 5.0

	// Indices 34-37: lead angle, closing speed, lead line-of-fire, distance
	if len(aobs) > 0 {
		nd := aobs[0] // nearest

		// 34: LEAD angle error — predict where nearest asteroid will be when
		// a bullet arrives, and report the angle to THAT position.
		t := nd.dist / bulletSpeed // approximate flight time
		leadDx := nd.dx + nd.avx*t
		leadDy := nd.dy + nd.avy*t
		leadAngle := math.Atan2(leadDy, leadDx)
		obs[34] = normalizeAngle(leadAngle-angle) / math.Pi

		// 35: Closing speed
		if nd.dist > 0 {
			dirX := nd.dx / nd.dist
			dirY := nd.dy / nd.dist
			obs[35] = (vx*dirX + vy*dirY) / maxSpeed
		}

		// 37: Distance to nearest asteroid
		obs[37] = nd.dist / halfDiag

		// 36: Lead line-of-fire quality — best shot quality using PREDICTED
		// positions. Accounts for asteroid movement during bullet flight.
		bestShot := 0.0
		for _, a := range aobs {
			if a.dist < 1 {
				continue
			}
			ft := a.dist / bulletSpeed
			pdx := a.dx + a.avx*ft
			pdy := a.dy + a.avy*ft
			pdist := math.Sqrt(pdx*pdx + pdy*pdy)
			if pdist < 1 {
				bestShot = 1.0
				continue
			}
			coneHalf := math.Atan2(a.radius, pdist)
			angleToA := math.Atan2(pdy, pdx)
			errA := math.Abs(normalizeAngle(angleToA - angle))
			if errA < coneHalf {
				q := 1.0 - errA/coneHalf
				if q > bestShot {
					bestShot = q
				}
			}
		}
		obs[36] = bestShot
	}

	// Indices 38-39: collision threat detection
	// Check all asteroids AND saucer bullets for collision courses.
	// Uses closest-point-of-approach (CPA) to detect "this thing is heading
	// straight for me."
	bestUrgency := 0.0
	dodgeDx, dodgeDy := 0.0, 0.0

	// Check asteroids
	for _, a := range aobs {
		// Relative velocity of asteroid toward player
		rvx, rvy := a.dvx, a.dvy
		rvSq := rvx*rvx + rvy*rvy
		if rvSq < 0.01 {
			continue // barely moving relative to player
		}
		// Time of closest approach
		tCPA := -(a.dx*rvx + a.dy*rvy) / rvSq
		if tCPA < 0 || tCPA > 120 { // past or too far in future (2 sec)
			continue
		}
		// Miss distance at CPA
		cx := a.dx + rvx*tCPA
		cy := a.dy + rvy*tCPA
		missDist := math.Sqrt(cx*cx + cy*cy)
		threatRadius := (playerRadius + a.radius) * 2 // generous margin
		if missDist < threatRadius {
			urgency := 1.0 / (1.0 + tCPA/30.0) // 30 frames ≈ 0.5 sec
			if urgency > bestUrgency {
				bestUrgency = urgency
				// Dodge direction: perpendicular to the relative velocity,
				// on the side that takes us AWAY from the asteroid.
				// Cross product sign tells us which side.
				cross := a.dx*rvy - a.dy*rvx
				if cross >= 0 {
					dodgeDx, dodgeDy = -rvy, rvx
				} else {
					dodgeDx, dodgeDy = rvy, -rvx
				}
			}
		}
	}

	// Check saucer bullets (same CPA logic)
	for _, sb := range sBullets {
		sdx := WrapDelta(px, sb.X, ScreenWidth)
		sdy := WrapDelta(py, sb.Y, ScreenHeight)
		srvx := sb.VX - vx
		srvy := sb.VY - vy
		rvSq := srvx*srvx + srvy*srvy
		if rvSq < 0.01 {
			continue
		}
		tCPA := -(sdx*srvx + sdy*srvy) / rvSq
		if tCPA < 0 || tCPA > 90 {
			continue
		}
		cx := sdx + srvx*tCPA
		cy := sdy + srvy*tCPA
		missDist := math.Sqrt(cx*cx + cy*cy)
		if missDist < playerRadius*3 {
			urgency := 1.0 / (1.0 + tCPA/20.0)
			if urgency > bestUrgency {
				bestUrgency = urgency
				cross := sdx*srvy - sdy*srvx
				if cross >= 0 {
					dodgeDx, dodgeDy = -srvy, srvx
				} else {
					dodgeDx, dodgeDy = srvy, -srvx
				}
			}
		}
	}

	obs[38] = bestUrgency
	if bestUrgency > 0 {
		dodgeAngle := math.Atan2(dodgeDy, dodgeDx)
		obs[39] = normalizeAngle(dodgeAngle-angle) / math.Pi
	}

	return obs
}

// normalizeAngle wraps an angle to [-π, π].
func normalizeAngle(a float64) float64 {
	for a > math.Pi {
		a -= 2 * math.Pi
	}
	for a < -math.Pi {
		a += 2 * math.Pi
	}
	return a
}

// --- AI fitness helpers ---

// NearestAsteroidAim returns the cosine of the angle between the player's
// facing direction and the direction to the nearest asteroid, plus the
// distance to that asteroid. Returns (0, 0) if no asteroids exist.
// aimCos ranges [-1, 1]: +1 means perfectly aimed, -1 means facing away.
func NearestAsteroidAim(w *World) (aimCos, dist float64) {
	px, py := w.PlayerPosition()
	angle := w.PlayerAngle()
	facingX := math.Cos(angle)
	facingY := math.Sin(angle)

	bestDist := math.Inf(1)
	bestDx, bestDy := 0.0, 0.0

	for e := range w.asteroids {
		pos := w.positions[e]
		if pos == nil {
			continue
		}
		dx := WrapDelta(px, pos.X, ScreenWidth)
		dy := WrapDelta(py, pos.Y, ScreenHeight)
		d := math.Sqrt(dx*dx + dy*dy)
		if d < bestDist {
			bestDist = d
			bestDx, bestDy = dx, dy
		}
	}

	if math.IsInf(bestDist, 1) {
		return 0, 0
	}

	// Normalize direction to nearest asteroid
	dirX := bestDx / bestDist
	dirY := bestDy / bestDist

	// Dot product = cos(angle between facing and target)
	return facingX*dirX + facingY*dirY, bestDist
}

// --- Exported spawn helper ---

// SpawnInitialWave is an exported wrapper around spawnWave for use by the AI
// training loop.
func SpawnInitialWave(w *World) {
	spawnWave(w)
}
