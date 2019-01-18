defmodule ArchitectWeb.Mutations.KnownHostsMutationsTest do
  alias Architect.KnownHosts
  alias Architect.KnownHosts.KnownHost
  use ArchitectWeb.ConnCase
  import Kronky.TestHelper
  alias Kronky.ValidationMessage
  alias Architect.Repo
  @tag :slow

  @known_hosts [
    %KnownHost{
      host: "bitbucket.com"
    }
  ]

  @fields %{
    id: :string,
    entry: :string,
    host: :string,
    fingerprint_md5: :string,
    fingerprint_sha256: :string,
    verified: :boolean
  }

  setup do
    known_hosts = for known_host <- @known_hosts, do: Repo.insert!(known_host)

    [known_hosts: known_hosts]
  end

  describe "createKnownHost" do
    test "Success", %{} do
      mutation = "
        mutation {
          createKnownHost(host: \"github.com\") {
            result {
              id,
              entry,
              host,
              fingerprintMd5,
              fingerprintSha256,
              verified
            },
            successful,
            messages {
              field
              message
            }
          }
        }
      "

      %{"createKnownHost" => actual} =
        graphql_request(mutation)
        |> expect_success!()

      expected = KnownHosts.get_known_host_by_host!("github.com")

      assert_mutation_success(expected, actual, @fields)
    end

    test "Failure - invalid host", %{} do
      mutation = "
        mutation {
          createKnownHost(host: \"invalid\") {
            result {
              id
            },
            successful,
            messages {
              field,
              code,
              message
            }
          }
        }
        "

      %{"createKnownHost" => actual} =
        graphql_request(mutation)
        |> expect_success!()

      expected = %ValidationMessage{
        code: :unknown,
        field: :host,
        message: "Scanning host failed"
      }

      assert_mutation_failure([expected], actual, [:code])
    end

    test "Failure - Unauthorized", %{} do
      mutation = "
        mutation {
          createKnownHost(host: \"github.com\") {
            result {
              id
            }
          }
        }
      "

      unauthorized_graphql_request(mutation)
      |> expect_failure!()
      |> Enum.find(fn %{"message" => message} -> message == "Unauthorized" end)
      |> assert
    end
  end

  describe "verifyKnownHost" do
    test "Success", %{known_hosts: [unverified | _]} do
      mutation = "
        mutation {
          verifyKnownHost(id: \"#{unverified.id}\") {
            result {
              id,
              host,
              verified
            },
            successful,
            messages {
              field
              message
            }
          }
        }
      "

      %{"verifyKnownHost" => actual} =
        graphql_request(mutation)
        |> expect_success!()

      expected = %KnownHost{unverified | verified: true}

      assert_mutation_success(expected, actual, @fields)
    end

    test "Failure - Unauthorized", %{known_hosts: [unverified | _]} do
      mutation = "
        mutation {
          verifyKnownHost(id: \"#{unverified.id}\") {
            result {
              id,
              host,
              verified
            },
            successful,
            messages {
              field
              message
            }
          }
        }
      "

      unauthorized_graphql_request(mutation)
      |> expect_failure!()
      |> Enum.find(fn %{"message" => message} -> message == "Unauthorized" end)
      |> assert
    end
  end
end
