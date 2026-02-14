package game

// Entity is a unique identifier for a game object.
type Entity uint64

// World holds all entities and their component stores.
type World struct {
	nextID   Entity
	entities map[Entity]bool

	// Component stores (struct-of-arrays style)
	positions     map[Entity]*Position
	velocities    map[Entity]*Velocity
	rotations     map[Entity]*Rotation
	colliders     map[Entity]*Collider
	renderables   map[Entity]*Renderable
	players       map[Entity]*PlayerControl
	asteroids     map[Entity]*AsteroidTag
	bullets       map[Entity]*BulletTag
	particles     map[Entity]*ParticleTag
	saucers       map[Entity]*SaucerTag
	saucerBullets map[Entity]*SaucerBulletTag
	wrappers      map[Entity]bool // entities that wrap around screen
}

func NewWorld() *World {
	return &World{
		nextID:        1,
		entities:      make(map[Entity]bool),
		positions:     make(map[Entity]*Position),
		velocities:    make(map[Entity]*Velocity),
		rotations:     make(map[Entity]*Rotation),
		colliders:     make(map[Entity]*Collider),
		renderables:   make(map[Entity]*Renderable),
		players:       make(map[Entity]*PlayerControl),
		asteroids:     make(map[Entity]*AsteroidTag),
		bullets:       make(map[Entity]*BulletTag),
		particles:     make(map[Entity]*ParticleTag),
		saucers:       make(map[Entity]*SaucerTag),
		saucerBullets: make(map[Entity]*SaucerBulletTag),
		wrappers:      make(map[Entity]bool),
	}
}

func (w *World) Spawn() Entity {
	id := w.nextID
	w.nextID++
	w.entities[id] = true
	return id
}

func (w *World) Destroy(e Entity) {
	delete(w.entities, e)
	delete(w.positions, e)
	delete(w.velocities, e)
	delete(w.rotations, e)
	delete(w.colliders, e)
	delete(w.renderables, e)
	delete(w.players, e)
	delete(w.asteroids, e)
	delete(w.bullets, e)
	delete(w.particles, e)
	delete(w.saucers, e)
	delete(w.saucerBullets, e)
	delete(w.wrappers, e)
}

// Alive returns whether an entity still exists.
func (w *World) Alive(e Entity) bool {
	return w.entities[e]
}

// BulletCount returns the number of active player bullets.
func (w *World) BulletCount() int {
	return len(w.bullets)
}
