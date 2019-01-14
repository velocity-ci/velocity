defmodule ArchitectWeb.Queries.KnownHostsTest do
  alias Architect.KnownHosts.KnownHost
  use ArchitectWeb.ConnCase
  import Kronky.TestHelper
  alias Architect.Repo

  @known_hosts [
    %KnownHost{
      entry: "github.com ssh-rsa ...",
      host: "github.com",
      fingerprint_sha256: "SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8",
      fingerprint_md5: "16:27:ac:a5:76:28:2d:36:63:1b:56:4d:eb:df:a6:48",
      verified: true
    },
    %KnownHost{
      entry: "bitbucket.com ssh-rsa ...",
      host: "bitbucket.com",
      fingerprint_sha256: "SHA256:zzXQOXSRBEiUtuE8AikJYKwbHaxvSc0ojez9YXaGp1A",
      fingerprint_md5: "97:8c:1b:f2:6f:14:6b:5c:3b:ec:aa:46:46:74:7c:40",
      verified: false
    }
  ]

  @fields %{
    id: :string,
    host: :string,
    entry: :string,
    fingerprint_md5: :string,
    fingerprint_sha256: :string,
    verified: :boolean
  }

  setup do
    known_hosts = for known_host <- @known_hosts, do: Repo.insert!(known_host)

    [conn: build_conn(), known_hosts: known_hosts]
  end

  test "gets a list of all known hosts", %{conn: conn, known_hosts: known_hosts} do
    query = """
      {
        knownHosts {
          id,
          host,
          entry,
          fingerprintMd5,
          fingerprintSha256,
          verified
        }
      }
    """

    %{"data" => %{"knownHosts" => actual}} =
      conn
      |> post("/v2", %{query: query})
      |> json_response(200)

    assert_equivalent_graphql(known_hosts, actual, @fields)
  end
end
