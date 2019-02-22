package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Worker struct {
	s       *Storage
	l       *pq.Listener
	closeCh chan struct{}
	outCh   chan struct{}
}

func (s *Storage) NewWorkerConnection() *Worker {
	w := &Worker{
		s: s,
		l: pq.NewListener(s.pgURL, 2*time.Second, 10*time.Second, func(ev pq.ListenerEventType, err error) {
			if err != nil {
				s.errLog.Println(err)
			}
		}),
		closeCh: make(chan struct{}),
		outCh:   make(chan struct{}),
	}
	go w.run()
	return w
}

func (w *Worker) run() {
	for {
		if err := w.l.Listen("jobs_notify"); err != nil {
			w.s.errLog.Println(err)
		} else {
			break
		}
	}
	n := w.l.NotificationChannel()
	for {
		select {
		case <-w.closeCh:
			return
		case notif := <-n:
			if notif != nil {
				select {
				case <-w.closeCh:
					return
				case w.outCh <- struct{}{}:
				}
			}
		}
	}
}

func (w *Worker) NotifCh() <-chan struct{} {
	return w.outCh
}

func (w *Worker) Close() error {
	close(w.closeCh)
	time.Sleep(10 * time.Millisecond)
	close(w.outCh)
	return w.l.Close()
}

func (w *Worker) PostJob(ctx context.Context, data int64) (int64, error) {
	var id int64
	if err := w.s.db.QueryRowContext(ctx, ``+
		`insert into jobs(input)`+
		`values ($1)`+
		`returning id;`,
		data,
	).Scan(&id); err != nil {
		w.s.errLog.Println(err)
		return 0, err

	}

	return id, nil
}

func (w *Worker) TakeJob(ctx context.Context) (id int64, data int64, err error) {
	tx, err := w.s.db.BeginTx(ctx, nil)
	if err != nil {
		w.s.errLog.Println(err)
		return 0, 0, err
	}
	commited := false
	defer func() {
		if !commited {
			tx.Rollback()
		}
	}()

	if err := tx.QueryRowContext(ctx, ``+
		`lock table jobs in access exclusive mode; `+
		`select id, input from jobs where process is null and complete is null limit 1;`,
	).Scan(&id, &data); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, ErrNotFound
		}
		w.s.errLog.Println(err)
		return 0, 0, err
	}

	if _, err := tx.ExecContext(ctx, ``+
		`update jobs set process = now() where id = $1`,
		id,
	); err != nil {
		w.s.errLog.Println(err)
		return 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		w.s.errLog.Println(err)
		return 0, 0, err
	}

	commited = true
	return id, data, nil
}

func (w *Worker) UntakeJob(ctx context.Context, id int64) error {
	if _, err := w.s.db.ExecContext(ctx, ``+
		`update jobs set process = null `+
		`where id = $1 and process is not null and complete is null`,
		id,
	); err != nil {
		w.s.errLog.Println(err)
		return err
	}

	return nil
}

func (w *Worker) PublishJobResult(ctx context.Context, id int64, data int64) error {
	if _, err := w.s.db.ExecContext(ctx, ``+
		`update jobs set output = $1, complete = now() `+
		`where id = $2 and process is not null and complete is null;`,
		data, id,
	); err != nil {
		w.s.errLog.Println(err)
		return err
	}

	return nil
}

func (w *Worker) GetJobResult(ctx context.Context, id int64) (score int64, err error) {
	if err := w.s.db.QueryRowContext(ctx, ``+
		`select output from jobs `+
		`where id = $1 and process is not null and complete is not null`,
		id,
	).Scan(&score); err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNotFound
		}
		w.s.errLog.Println(err)
		return 0, err
	}

	return
}
