package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"task250-biosonar/internal/model"
	"task250-biosonar/internal/service"
	"task250-biosonar/internal/store"
)

// runSmokeTest exercises the full lifecycle against a temporary database, then
// closes and reopens it to prove persistence and restart recovery. It returns
// nil only when every invariant holds after restart.
func runSmokeTest() error {
	dir, err := os.MkdirTemp("", "biosonar-smoke-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	dbPath := filepath.Join(dir, "smoke.db")

	st, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	svc := service.New(st)

	// 1. create a survey batch.
	b := &model.SurveyBatch{Name: "smoke-line-1", Vessel: "RV Test"}
	bid, err := st.CreateBatch(b)
	if err != nil {
		return fmt.Errorf("create batch: %w", err)
	}

	// 2. ingest 12 pings: sand-like (0-4), an attitude anomaly (5), mud-like
	//    (6-11).
	base := time.Now().UTC()
	for i := 0; i < 12; i++ {
		att := model.Attitude{Pitch: 1.0, Roll: 1.0, Heading: 0, Heave: 0.1}
		if i == 5 {
			att.Pitch = 20.0
			att.Roll = 18.0
		}
		var ch []model.EchoChannel
		if i < 6 {
			ch = sandChannels()
		} else {
			ch = mudChannels()
		}
		e := &model.EchoWindow{
			BatchID:       bid,
			PingSeq:       i,
			PosX:          float64(i * 10),
			PosY:          0,
			Timestamp:     base.Add(time.Duration(i) * time.Second),
			Attitude:      att,
			SoundVelocity: 1500,
			SlantRange:    50.0,
			Channels:      ch,
			Status:        model.EchoRaw,
		}
		if _, err := svc.IngestEcho(e); err != nil {
			return fmt.Errorf("ingest %d: %w", i, err)
		}
	}

	// 3. classify every usable ping.
	echoes, err := st.ListEchoes(bid)
	if err != nil {
		return fmt.Errorf("list echoes: %w", err)
	}
	for _, e := range echoes {
		if e.Status == model.EchoAttitudeAnom {
			continue
		}
		if _, err := svc.ClassifyEcho(e.ID); err != nil {
			return fmt.Errorf("classify %d: %w", e.ID, err)
		}
	}

	// 4. merge spatially contiguous runs into segments.
	segs, err := svc.MergeBatch(bid)
	if err != nil {
		return fmt.Errorf("merge: %w", err)
	}
	if len(segs) == 0 {
		return fmt.Errorf("expected segments, got none")
	}

	// 5. publish an interpretation snapshot.
	snap, err := svc.PublishBatchSnapshot(bid, "smoke")
	if err != nil {
		return fmt.Errorf("snapshot: %w", err)
	}
	if snap.Status != model.SnapPublished {
		return fmt.Errorf("snapshot not published: %s", snap.Status)
	}

	// 6. drive the batch through its full state machine to sealed.
	for _, to := range []model.BatchStatus{
		model.BatchPendingCorr, model.BatchPendingClass, model.BatchPublished, model.BatchSealed,
	} {
		if _, err := st.TransitionBatch(bid, to); err != nil {
			return fmt.Errorf("transition %s: %w", to, err)
		}
	}

	// 7. close and reopen to verify persistence + restart recovery.
	if err := st.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	st2, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("reopen: %w", err)
	}
	defer st2.Close()

	b2, err := st2.GetBatch(bid)
	if err != nil {
		return fmt.Errorf("reopen batch: %w", err)
	}
	if b2.Status != model.BatchSealed {
		return fmt.Errorf("reopen status mismatch: %s", b2.Status)
	}
	echoes2, err := st2.ListEchoes(bid)
	if err != nil {
		return fmt.Errorf("reopen echoes: %w", err)
	}
	if len(echoes2) != 12 {
		return fmt.Errorf("reopen echo count mismatch: %d", len(echoes2))
	}
	segs2, err := st2.ListSegments(bid)
	if err != nil {
		return fmt.Errorf("reopen segments: %w", err)
	}
	if len(segs2) == 0 {
		return fmt.Errorf("reopen lost segments")
	}
	snap2, err := st2.GetSnapshot(snap.ID)
	if err != nil {
		return fmt.Errorf("reopen snapshot: %w", err)
	}
	if snap2.Status != model.SnapPublished {
		return fmt.Errorf("reopen snapshot status: %s", snap2.Status)
	}
	return nil
}

// buildChannels synthesises a multispectral echo return whose backscatter
// decays with depth and drops below the penetration threshold at the given
// depth, yielding sand- or mud-like features.
func buildChannels(peakDB, penetration float64) []model.EchoChannel {
	freqs := []float64{120000, 200000, 400000}
	inc := []float64{20, 30, 40}
	depths := make([]float64, 21)
	for i := range depths {
		depths[i] = float64(i) * 0.5
	}
	chans := make([]model.EchoChannel, 3)
	for k := range freqs {
		amps := make([]float64, len(depths))
		for i, d := range depths {
			a := peakDB - 0.3*d
			if d > penetration {
				a -= 4.0
			}
			amps[i] = a
		}
		chans[k] = model.EchoChannel{
			FrequencyHz:  freqs[k],
			IncidenceDeg: inc[k],
			Depths:       depths,
			Amplitudes:   amps,
		}
	}
	return chans
}

func sandChannels() []model.EchoChannel { return buildChannels(-10.0, 2.5) }
func mudChannels() []model.EchoChannel  { return buildChannels(-20.0, 8.0) }
