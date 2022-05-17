package movr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/faker"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"golang.org/x/exp/rand"
)

type rideInfo struct {
	id   string
	city string
}

type movrWorker struct {
	db           *workload.RoundRobinDB
	hists        *histogram.Histograms
	activeRides  []rideInfo
	rng          *rand.Rand
	faker        faker.Faker
	creationTime time.Time
}

func (m *movrWorker) getRandomUser(city string) (string, error) {
	__antithesis_instrumentation__.Notify(694875)
	id, err := uuid.NewV4()
	if err != nil {
		__antithesis_instrumentation__.Notify(694877)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(694878)
	}
	__antithesis_instrumentation__.Notify(694876)
	var user string
	q := `
		SELECT
			IFNULL(a, b)
		FROM
			(
				SELECT
					(SELECT id FROM users WHERE city = $1 AND id > $2 ORDER BY id LIMIT 1)
						AS a,
					(SELECT id FROM users WHERE city = $1 ORDER BY id LIMIT 1) AS b
			);
		`
	err = m.db.QueryRow(q, city, id.String()).Scan(&user)
	return user, err
}

func (m *movrWorker) getRandomPromoCode() (string, error) {
	__antithesis_instrumentation__.Notify(694879)
	id, err := uuid.NewV4()
	if err != nil {
		__antithesis_instrumentation__.Notify(694881)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(694882)
	}
	__antithesis_instrumentation__.Notify(694880)
	q := `
		SELECT
			IFNULL(a, b)
		FROM
			(
				SELECT
					(SELECT code FROM promo_codes WHERE code > $1 ORDER BY code LIMIT 1)
						AS a,
					(SELECT code FROM promo_codes ORDER BY code LIMIT 1) AS b
			);
		`
	var code string
	err = m.db.QueryRow(q, id.String()).Scan(&code)
	return code, err
}

func (m *movrWorker) getRandomVehicle(city string) (string, error) {
	__antithesis_instrumentation__.Notify(694883)
	id, err := uuid.NewV4()
	if err != nil {
		__antithesis_instrumentation__.Notify(694885)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(694886)
	}
	__antithesis_instrumentation__.Notify(694884)
	q := `
		SELECT
			IFNULL(a, b)
		FROM
			(
				SELECT
					(SELECT id FROM vehicles WHERE city = $1 AND id > $2 ORDER BY id LIMIT 1)
						AS a,
					(SELECT id FROM vehicles WHERE city = $1 ORDER BY id LIMIT 1) AS b
			);
		`
	var vehicle string
	err = m.db.QueryRow(q, city, id.String()).Scan(&vehicle)
	return vehicle, err
}

func (m *movrWorker) readVehicles(city string) error {
	__antithesis_instrumentation__.Notify(694887)
	q := `SELECT city, id FROM vehicles WHERE city = $1`
	_, err := m.db.Exec(q, city)
	return err
}

func (m *movrWorker) updateActiveRides() error {
	__antithesis_instrumentation__.Notify(694888)
	for i, ride := range m.activeRides {
		__antithesis_instrumentation__.Notify(694890)
		if i >= 10 {
			__antithesis_instrumentation__.Notify(694892)
			break
		} else {
			__antithesis_instrumentation__.Notify(694893)
		}
		__antithesis_instrumentation__.Notify(694891)
		lat, long := randLatLong(m.rng)
		q := `UPSERT INTO vehicle_location_histories VALUES ($1, $2, now(), $3, $4)`
		_, err := m.db.Exec(q, ride.city, ride.id, lat, long)
		if err != nil {
			__antithesis_instrumentation__.Notify(694894)
			return err
		} else {
			__antithesis_instrumentation__.Notify(694895)
		}
	}
	__antithesis_instrumentation__.Notify(694889)
	return nil
}

func (m *movrWorker) addUser(id uuid.UUID, city string) error {
	__antithesis_instrumentation__.Notify(694896)
	q := `INSERT INTO users VALUES ($1, $2, $3, $4, $5)`
	_, err := m.db.Exec(
		q, id.String(), city, m.faker.Name(m.rng), m.faker.StreetAddress(m.rng), randCreditCard(m.rng))
	return err
}

func (m *movrWorker) createPromoCode(id uuid.UUID, _ string) error {
	__antithesis_instrumentation__.Notify(694897)
	expirationTime := m.creationTime.Add(time.Duration(m.rng.Intn(30)) * 24 * time.Hour)
	creationTime := expirationTime.Add(-time.Duration(m.rng.Intn(30)) * 24 * time.Hour)
	const rulesJSON = `{"type": "percent_discount", "value": "10%"}`
	q := `INSERT INTO promo_codes VALUES ($1, $2, $3, $4, $5)`
	_, err := m.db.Exec(q, id.String(), m.faker.Paragraph(m.rng), creationTime, expirationTime, rulesJSON)
	return err
}

func (m *movrWorker) applyPromoCode(id uuid.UUID, city string) error {
	__antithesis_instrumentation__.Notify(694898)
	user, err := m.getRandomUser(city)
	if err != nil {
		__antithesis_instrumentation__.Notify(694903)
		return err
	} else {
		__antithesis_instrumentation__.Notify(694904)
	}
	__antithesis_instrumentation__.Notify(694899)
	code, err := m.getRandomPromoCode()
	if err != nil {
		__antithesis_instrumentation__.Notify(694905)
		return err
	} else {
		__antithesis_instrumentation__.Notify(694906)
	}
	__antithesis_instrumentation__.Notify(694900)

	var count int
	q := `SELECT count(*) FROM user_promo_codes WHERE city = $1 AND user_id = $2 AND code = $3`
	err = m.db.QueryRow(q, city, user, code).Scan(&count)
	if err != nil {
		__antithesis_instrumentation__.Notify(694907)
		return err
	} else {
		__antithesis_instrumentation__.Notify(694908)
	}
	__antithesis_instrumentation__.Notify(694901)

	if count == 0 {
		__antithesis_instrumentation__.Notify(694909)
		q = `INSERT INTO user_promo_codes VALUES ($1, $2, $3, now(), 1)`
		_, err = m.db.Exec(q, city, user, code)
		return err
	} else {
		__antithesis_instrumentation__.Notify(694910)
	}
	__antithesis_instrumentation__.Notify(694902)
	return nil
}

func (m *movrWorker) addVehicle(id uuid.UUID, city string) error {
	__antithesis_instrumentation__.Notify(694911)
	ownerID, err := m.getRandomUser(city)
	if err != nil {
		__antithesis_instrumentation__.Notify(694913)
		return err
	} else {
		__antithesis_instrumentation__.Notify(694914)
	}
	__antithesis_instrumentation__.Notify(694912)
	typ := randVehicleType(m.rng)
	q := `INSERT INTO vehicles VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = m.db.Exec(
		q, id.String(), city, typ, ownerID,
		m.creationTime.Format(timestampFormat),
		randVehicleStatus(m.rng),
		m.faker.StreetAddress(m.rng),
		randVehicleMetadata(m.rng, typ),
	)
	return err
}

func (m *movrWorker) startRide(id uuid.UUID, city string) error {
	__antithesis_instrumentation__.Notify(694915)
	rider, err := m.getRandomUser(city)
	if err != nil {
		__antithesis_instrumentation__.Notify(694919)
		return err
	} else {
		__antithesis_instrumentation__.Notify(694920)
	}
	__antithesis_instrumentation__.Notify(694916)
	vehicle, err := m.getRandomVehicle(city)
	if err != nil {
		__antithesis_instrumentation__.Notify(694921)
		return err
	} else {
		__antithesis_instrumentation__.Notify(694922)
	}
	__antithesis_instrumentation__.Notify(694917)
	q := `INSERT INTO rides VALUES ($1, $2, $2, $3, $4, $5, NULL, now(), NULL, $6)`
	_, err = m.db.Exec(q, id.String(), city, rider, vehicle, m.faker.StreetAddress(m.rng), m.rng.Intn(100))
	if err != nil {
		__antithesis_instrumentation__.Notify(694923)
		return err
	} else {
		__antithesis_instrumentation__.Notify(694924)
	}
	__antithesis_instrumentation__.Notify(694918)
	m.activeRides = append(m.activeRides, rideInfo{id.String(), city})
	return err
}

func (m *movrWorker) endRide(id uuid.UUID, city string) error {
	__antithesis_instrumentation__.Notify(694925)
	if len(m.activeRides) > 1 {
		__antithesis_instrumentation__.Notify(694927)
		ride := m.activeRides[0]
		m.activeRides = m.activeRides[1:]
		q := `UPDATE rides SET end_address = $3, end_time = now() WHERE city = $1 AND id = $2`
		_, err := m.db.Exec(q, ride.city, ride.id, m.faker.StreetAddress(m.rng))
		return err
	} else {
		__antithesis_instrumentation__.Notify(694928)
	}
	__antithesis_instrumentation__.Notify(694926)
	return nil
}

func (m *movrWorker) generateWorkSimulation() func(context.Context) error {
	__antithesis_instrumentation__.Notify(694929)
	const readPercentage = 0.95
	movrWorkloadFns := []struct {
		weight float32
		key    string
		work   func(uuid.UUID, string) error
	}{
		{
			weight: 0.03,
			key:    "createPromoCode",
			work:   m.createPromoCode,
		},
		{
			weight: 0.1,
			key:    "applyPromoCode",
			work:   m.applyPromoCode,
		},
		{
			weight: 0.3,
			key:    "addUser",
			work:   m.addUser,
		},
		{
			weight: 0.1,
			key:    "addVehicle",
			work:   m.addVehicle,
		},
		{
			weight: 0.4,
			key:    "startRide",
			work:   m.startRide,
		},
		{
			weight: 0.07,
			key:    "endRide",
			work:   m.endRide,
		},
	}

	sum := float32(0.0)
	for _, s := range movrWorkloadFns {
		__antithesis_instrumentation__.Notify(694932)
		sum += s.weight
	}
	__antithesis_instrumentation__.Notify(694930)

	runAndRecord := func(key string, work func() error) error {
		__antithesis_instrumentation__.Notify(694933)
		start := timeutil.Now()
		err := work()
		elapsed := timeutil.Since(start)
		if err == nil {
			__antithesis_instrumentation__.Notify(694935)
			m.hists.Get(key).Record(elapsed)
		} else {
			__antithesis_instrumentation__.Notify(694936)
		}
		__antithesis_instrumentation__.Notify(694934)
		return err
	}
	__antithesis_instrumentation__.Notify(694931)

	return func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(694937)
		activeCity := randCity(m.rng)
		id, err := uuid.NewV4()
		if err != nil {
			__antithesis_instrumentation__.Notify(694943)
			return err
		} else {
			__antithesis_instrumentation__.Notify(694944)
		}
		__antithesis_instrumentation__.Notify(694938)

		if m.rng.Float64() <= readPercentage {
			__antithesis_instrumentation__.Notify(694945)
			return runAndRecord("readVehicles", func() error {
				__antithesis_instrumentation__.Notify(694946)
				return m.readVehicles(activeCity)
			})
		} else {
			__antithesis_instrumentation__.Notify(694947)
		}
		__antithesis_instrumentation__.Notify(694939)
		err = runAndRecord("updateActiveRides", func() error {
			__antithesis_instrumentation__.Notify(694948)
			return m.updateActiveRides()
		})
		__antithesis_instrumentation__.Notify(694940)
		if err != nil {
			__antithesis_instrumentation__.Notify(694949)
			return err
		} else {
			__antithesis_instrumentation__.Notify(694950)
		}
		__antithesis_instrumentation__.Notify(694941)
		randVal := m.rng.Float32() * sum
		w := float32(0.0)
		for _, s := range movrWorkloadFns {
			__antithesis_instrumentation__.Notify(694951)
			w += s.weight
			if w >= randVal {
				__antithesis_instrumentation__.Notify(694952)
				return runAndRecord(s.key, func() error {
					__antithesis_instrumentation__.Notify(694953)
					return s.work(id, activeCity)
				})
			} else {
				__antithesis_instrumentation__.Notify(694954)
			}
		}
		__antithesis_instrumentation__.Notify(694942)
		panic("unreachable")
	}
}

func (m *movr) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(694955)

	m.fakerOnce.Do(func() {
		__antithesis_instrumentation__.Notify(694959)
		m.faker = faker.NewFaker()
	})
	__antithesis_instrumentation__.Notify(694956)
	sqlDatabase, err := workload.SanitizeUrls(m, m.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(694960)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(694961)
	}
	__antithesis_instrumentation__.Notify(694957)
	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	db, err := workload.NewRoundRobinDB(urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(694962)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(694963)
	}
	__antithesis_instrumentation__.Notify(694958)
	worker := movrWorker{
		db:           db,
		rng:          rand.New(rand.NewSource(m.seed)),
		faker:        m.faker,
		creationTime: m.creationTime,
		activeRides:  []rideInfo{},
		hists:        reg.GetHandle(),
	}
	ql.WorkerFns = append(ql.WorkerFns, worker.generateWorkSimulation())

	return ql, nil
}
