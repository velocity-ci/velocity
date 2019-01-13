defmodule Architect.KnownHostsTest do
  use Architect.DataCase

  alias Architect.KnownHosts

  @tag :slow

  describe "known_hosts" do
    alias Architect.KnownHosts.KnownHost

    @valid_attrs %{
      host: "github.com"
    }

    @invalid_attrs %{
      verified: "not a bool"
    }

    @update_attrs %{
      verified: true
    }

    def known_host_fixture(attrs \\ %{}) do
      {:ok, known_host} =
        attrs
        |> Enum.into(@valid_attrs)
        |> KnownHosts.create_known_host()

      known_host
    end

    test "list_known_hosts/0 returns all known_hosts" do
      known_host = known_host_fixture()
      assert KnownHosts.list_known_hosts() == [known_host]
    end

    test "get_known_host!/1 returns the known_host with given id" do
      known_host = known_host_fixture()
      assert KnownHosts.get_known_host!(known_host.id) == known_host
    end

    test "create_known_host/1 with valid data creates a known_host" do
      assert {:ok, %KnownHost{} = known_host} = KnownHosts.create_known_host(@valid_attrs)

      assert known_host.entry ==
               "github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==\n"

      assert known_host.fingerprint_md5 == "16:27:ac:a5:76:28:2d:36:63:1b:56:4d:eb:df:a6:48"
      assert known_host.fingerprint_sha256 == "SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8"
      assert known_host.host == "github.com"
      refute known_host.verified
    end

    test "create_known_host/1 with invalid data returns error changeset" do
      assert {:error, %Ecto.Changeset{}} = KnownHosts.create_known_host(@invalid_attrs)
    end

    test "update_known_host/2 with valid data updates the known_host" do
      known_host = known_host_fixture()

      assert {:ok, %KnownHost{} = known_host} =
               KnownHosts.update_known_host(known_host, @update_attrs)

      assert known_host.verified
    end

    test "update_known_host/2 with invalid data returns error changeset" do
      known_host = known_host_fixture()

      assert {:error, %Ecto.Changeset{}} =
               KnownHosts.update_known_host(known_host, @invalid_attrs)

      assert known_host == KnownHosts.get_known_host!(known_host.id)
    end

    test "delete_known_host/1 deletes the known_host" do
      known_host = known_host_fixture()
      assert {:ok, %KnownHost{}} = KnownHosts.delete_known_host(known_host)
      assert_raise Ecto.NoResultsError, fn -> KnownHosts.get_known_host!(known_host.id) end
    end
  end
end
