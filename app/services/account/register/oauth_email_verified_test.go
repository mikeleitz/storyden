package register_test

import (
	"context"
	"net/mail"
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/authentication"
	"github.com/Southclaws/storyden/app/resources/account/email"
	"github.com/Southclaws/storyden/app/services/account/register"
	"github.com/Southclaws/storyden/internal/integration"
	"github.com/Southclaws/storyden/internal/integration/e2e"
)

func TestGetOrCreateViaEmailProviderVerified(t *testing.T) {
	t.Parallel()

	integration.Test(t, nil, e2e.Setup(), fx.Invoke(func(
		lc fx.Lifecycle,
		root context.Context,
		registrar *register.Registrar,
		emailRepo *email.Repository,
	) {
		lc.Append(fx.StartHook(func() {
			service := authentication.ServiceOAuthKeycloak
			token := "test-token"

			t.Run("new_account_trusts_provider_verification", func(t *testing.T) {
				r := require.New(t)
				a := assert.New(t)

				addr := uniqueEmail()
				acc, err := registrar.GetOrCreateViaEmail(
					root, service, "Keycloak", xid.New().String(), token,
					uniqueHandle(), "Verified User", addr, true,
				)
				r.NoError(err)
				r.NotNil(acc)

				a.True(emailVerified(t, emailRepo, root, addr, acc.ID))
			})

			t.Run("returning_login_trusts_provider_verification", func(t *testing.T) {
				r := require.New(t)
				a := assert.New(t)

				addr := uniqueEmail()
				handle := uniqueHandle()
				identifier := xid.New().String()

				acc1, err := registrar.GetOrCreateViaEmail(
					root, service, "Keycloak", identifier, token,
					handle, "Returning User", addr, false,
				)
				r.NoError(err)
				r.NotNil(acc1)
				a.False(emailVerified(t, emailRepo, root, addr, acc1.ID))

				acc2, err := registrar.GetOrCreateViaEmail(
					root, service, "Keycloak", identifier, token,
					handle, "Returning User", addr, true,
				)
				r.NoError(err)
				r.NotNil(acc2)
				a.Equal(acc1.ID, acc2.ID)
				a.True(emailVerified(t, emailRepo, root, addr, acc2.ID))
			})

			t.Run("existing_unverified_email_becomes_verified", func(t *testing.T) {
				r := require.New(t)
				a := assert.New(t)

				addr := uniqueEmail()
				handle := uniqueHandle()

				acc1, err := registrar.GetOrCreateViaEmail(
					root, service, "Keycloak", xid.New().String(), token,
					handle, "Pending User", addr, false,
				)
				r.NoError(err)
				r.NotNil(acc1)
				a.False(emailVerified(t, emailRepo, root, addr, acc1.ID))

				acc2, err := registrar.GetOrCreateViaEmail(
					root, service, "Keycloak", xid.New().String(), token,
					handle, "Pending User", addr, true,
				)
				r.NoError(err)
				r.NotNil(acc2)
				a.Equal(acc1.ID, acc2.ID)
				a.True(emailVerified(t, emailRepo, root, addr, acc2.ID))
			})

			t.Run("unverified_provider_cannot_claim_unverified_email", func(t *testing.T) {
				r := require.New(t)

				addr := uniqueEmail()
				handle := uniqueHandle()

				acc, err := registrar.GetOrCreateViaEmail(
					root, service, "Keycloak", xid.New().String(), token,
					handle, "Pending User", addr, false,
				)
				r.NoError(err)
				r.NotNil(acc)

				_, err = registrar.GetOrCreateViaEmail(
					root, service, "Keycloak", xid.New().String(), token,
					handle, "Other User", addr, false,
				)
				r.Error(err)
			})
		}))
	}))
}

func uniqueEmail() mail.Address {
	return mail.Address{Address: xid.New().String() + "@example.com"}
}

func uniqueHandle() string {
	return "h" + xid.New().String()
}

func emailVerified(
	t *testing.T,
	emailRepo *email.Repository,
	ctx context.Context,
	addr mail.Address,
	accID account.AccountID,
) bool {
	t.Helper()

	owner, exists, err := emailRepo.LookupAccount(ctx, addr)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, accID, owner.ID)

	for _, item := range owner.EmailAddresses {
		if item.Email.Address == addr.Address {
			return item.Verified
		}
	}

	t.Fatalf("email %s not found on account %s", addr.Address, accID)
	return false
}
