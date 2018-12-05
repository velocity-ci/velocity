defmodule Architect.KnownHostsTest do
  use Architect.DataCase

  alias Architect.KnownHosts

  describe "known_hosts" do
    alias Architect.KnownHosts.KnownHost

    @valid_attrs %{
      comment: "some comment",
      entry: "some entry",
      fingerprint_md5: "some fingerprint_md5",
      fingerprint_sha256: "some fingerprint_sha256",
      hosts: ["some hosts"]
    }
    @update_attrs %{
      comment: "some updated comment",
      entry: "some updated entry",
      fingerprint_md5: "some updated fingerprint_md5",
      fingerprint_sha256: "some updated fingerprint_sha256",
      hosts: ["some updated hosts"]
    }
    @invalid_attrs %{
      comment: nil,
      entry: nil,
      fingerprint_md5: nil,
      fingerprint_sha256: nil,
      hosts: [nil]
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
      assert known_host.comment == "some comment"
      assert known_host.entry == "some entry"
      assert known_host.fingerprint_md5 == "some fingerprint_md5"
      assert known_host.fingerprint_sha256 == "some fingerprint_sha256"
      assert known_host.hosts == ["some hosts"]
    end

    test "create_known_host/1 with invalid data returns error changeset" do
      assert {:error, %Ecto.Changeset{}} = KnownHosts.create_known_host(@invalid_attrs)
    end

    test "update_known_host/2 with valid data updates the known_host" do
      known_host = known_host_fixture()

      assert {:ok, %KnownHost{} = known_host} =
               KnownHosts.update_known_host(known_host, @update_attrs)

      assert known_host.comment == "some updated comment"
      assert known_host.entry == "some updated entry"
      assert known_host.fingerprint_md5 == "some updated fingerprint_md5"
      assert known_host.fingerprint_sha256 == "some updated fingerprint_sha256"
      assert known_host.hosts == ["some updated hosts"]
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

    test "change_known_host/1 returns a known_host changeset" do
      known_host = known_host_fixture()
      assert %Ecto.Changeset{} = KnownHosts.change_known_host(known_host)
    end
  end
end
