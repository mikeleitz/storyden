import { redirect } from "next/navigation";

import { authProviderList } from "@/api/openapi-server/auth";
import { Text } from "@/components/ui/text";

const BRIDGE_START =
  "https://daltons-office.easyeasyspeakeasy.com/realms/main/start";

const KEYCLOAK_PROVIDER = "oauth_keycloak";

type Props = {
  searchParams: Promise<{
    login_token?: string;
  }>;
};

export default async function MagicAuthPage(props: Props) {
  const { login_token } = await props.searchParams;
  if (!login_token) {
    return (
      <Text>
        Missing login token. Open the member dashboard via the SSH login link,
        then click Forum again within a few minutes.
      </Text>
    );
  }

  let providers;
  try {
    const { data } = await authProviderList({
      cache: "no-store",
    });
    providers = data.providers;
  } catch {
    return (
      <Text>
        Could not load sign-in providers. Check that the forum API is reachable.
      </Text>
    );
  }

  const keycloak = providers.find(
    (provider) => provider.provider === KEYCLOAK_PROVIDER && provider.link,
  );
  if (!keycloak?.link) {
    return (
      <Text>
        Keycloak sign-in is not enabled on this forum. Ask an administrator to
        enable the Keycloak OAuth provider.
      </Text>
    );
  }

  const start = new URL(BRIDGE_START);
  start.searchParams.set("login_token", login_token);
  start.searchParams.set("return_to", keycloak.link);

  redirect(start.toString());
}
