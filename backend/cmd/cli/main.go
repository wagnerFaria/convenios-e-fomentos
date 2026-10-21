// Comando cli reúne os comandos administrativos (migrações, administradores, etc.).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/wagner/convenios-e-fomentos/backend/internal/admin"
	"github.com/wagner/convenios-e-fomentos/backend/internal/app"
	"github.com/wagner/convenios-e-fomentos/backend/internal/config"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/logging"
	"github.com/wagner/convenios-e-fomentos/backend/internal/mailer"
	"github.com/wagner/convenios-e-fomentos/backend/internal/maintenance"
	"github.com/wagner/convenios-e-fomentos/backend/internal/seed"
)

const uso = `Uso: cli <comando> [opções]

Comandos:
  migrate:up         aplica as migrações pendentes
  migrate:down       desfaz a última migração
  migrate:status     mostra o estado das migrações
  admin:create       cria um Super Admin
                     --email E --password P [--first-name N] [--last-name S]
                     (a senha também pode vir da variável ADMIN_PASSWORD)
  seed               cria os dados iniciais (idempotente): roles e permissões padrão,
                     departamentos de exemplo, e, em desenvolvimento, um usuário de cada role
                     (as senhas aparecem no log somente na primeira execução)
  cleanup            remove sessões e tokens de uso único vencidos há mais de N dias
                     [--days 7] (agende, por exemplo, uma vez por dia)
  admin:reset-password
                     redefine a senha de um administrador e revoga as sessões
                     --email E --password P
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, uso)
		os.Exit(2)
	}
	if err := run(context.Background(), os.Args[1], os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cmd string, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	switch cmd {
	case "migrate:up", "migrate:down", "migrate:status":
		m := database.NewMigrator(pool)
		defer func() { _ = m.Close() }()
		switch cmd {
		case "migrate:up":
			return m.Up(ctx)
		case "migrate:down":
			return m.Down(ctx)
		default:
			return m.Status(ctx)
		}
	case "admin:create":
		fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
		email := fs.String("email", "", "e-mail do administrador")
		password := fs.String("password", os.Getenv("ADMIN_PASSWORD"), "senha")
		first := fs.String("first-name", "Super", "nome")
		last := fs.String("last-name", "Admin", "sobrenome")
		if err := fs.Parse(args); err != nil {
			return err
		}
		if *email == "" || *password == "" {
			return errors.New("--email e --password são obrigatórios")
		}
		a, err := app.AdminService(cfg, pool).CreateSuperAdmin(ctx, admin.CreateSuperAdminInput{FirstName: *first, LastName: *last, Email: *email, Password: *password}, false, "")
		if err != nil {
			return err
		}
		fmt.Printf("Super Admin criado: %s (%s)\n", a.Email, a.ID)
		return nil
	case "seed":
		a, err := app.New(cfg, logging.New(os.Stdout, cfg.LogLevel), pool, mailer.NopMailer{Log: logging.New(os.Stdout, cfg.LogLevel)})
		if err != nil {
			return err
		}
		return seed.Run(ctx, seed.Deps{Pool: pool, Hasher: app.Hasher(cfg), Engine: a.Authz, EndUsers: a.EndUsers, Log: logging.New(os.Stdout, cfg.LogLevel), Development: cfg.AppEnv == "development"})
	case "cleanup":
		fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
		days := fs.Int("days", 7, "carência, em dias, após o vencimento")
		if err := fs.Parse(args); err != nil {
			return err
		}
		res, err := maintenance.Cleanup(ctx, db.New(pool), *days)
		if err != nil {
			return err
		}
		fmt.Printf("Removidos: %d sessões de administrador, %d sessões de usuário final, %d tokens de uso único.\n", res.AdminSessions, res.EndUserSessions, res.OneTimeTokens)
		return nil
	case "admin:reset-password":
		fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
		email := fs.String("email", "", "e-mail do administrador")
		password := fs.String("password", os.Getenv("ADMIN_PASSWORD"), "nova senha")
		if err := fs.Parse(args); err != nil {
			return err
		}
		if *email == "" || *password == "" {
			return errors.New("--email e --password são obrigatórios")
		}
		if err := app.AdminService(cfg, pool).ResetPasswordByEmail(ctx, *email, *password); err != nil {
			return err
		}
		fmt.Println("Senha redefinida e sessões revogadas.")
		return nil
	default:
		return fmt.Errorf("comando desconhecido %q\n\n%s", cmd, uso)
	}
}
