package storage

import (
	"context"
	"database/sql"
)

func (s *Storage) AddUser(ctx context.Context, name string, pass string) error {
	if _, err := s.db.ExecContext(ctx, ``+
		`insert into users(name, pass)`+
		`values ($1, $2);`,
		name, pass,
	); err != nil {
		s.errLog.Println(err)
		return err
	}

	return nil
}

func (s *Storage) AddAdmin(ctx context.Context, name string, pass string) error {
	if _, err := s.db.ExecContext(ctx, ``+
		`insert into users(name, pass, is_admin)`+
		`values ($1, $2, true);`,
		name, pass,
	); err != nil {
		s.errLog.Println(err)
		return err
	}

	return nil
}

func (s *Storage) GetCounter(ctx context.Context, name string) (counter int64, err error) {
	if err = s.db.QueryRowContext(ctx, ``+
		`select counter from users where name = $1;`,
		name,
	).Scan(&counter); err != nil {
		s.errLog.Println(err)
		return 0, err
	}

	return
}

func (s *Storage) GetTotalCounter(ctx context.Context) (int64, error) {
	var total int64
	if err := s.db.QueryRowContext(ctx, ``+
		`select sum(counter) from users;`,
	).Scan(&total); err != nil {
		s.errLog.Println(err)
		return 0, err
	}

	return total, nil
}

func (s *Storage) IncCounter(ctx context.Context, name string) error {
	var counter int64

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.errLog.Println(err)
		return err
	}
	commited := false
	defer func() {
		if !commited {
			tx.Rollback()
		}
	}()

	if err := tx.QueryRowContext(ctx, ``+
		`select counter from users where name = $1;`,
		name,
	).Scan(&counter); err != nil {
		s.errLog.Println(err)
		return err
	}

	counter++

	if _, err := tx.ExecContext(ctx, ``+
		`update users set counter = $1 where name = $2`,
		counter, name,
	); err != nil {
		s.errLog.Println(err)
		return err
	}

	if err := tx.Commit(); err != nil {
		s.errLog.Println(err)
		return err
	}

	commited = true
	return nil
}

func (s *Storage) ResetCounter(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, ``+
		`update users set counter = 0`,
	); err != nil {
		s.errLog.Println(err)
		return err
	}

	return nil
}

func (s *Storage) IsAdmin(ctx context.Context, name string, pass string) (bool, error) {
	var isAdmin bool
	if err := s.db.QueryRowContext(ctx, ``+
		`select is_admin from users where name = $1 and pass = $2;`,
		name, pass,
	).Scan(&isAdmin); err != nil {
		if err == sql.ErrNoRows {
			return false, ErrNotFound
		}
		s.errLog.Println(err)
		return false, err
	}

	return isAdmin, nil
}
