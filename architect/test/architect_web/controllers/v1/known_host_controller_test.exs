defmodule ArchitectWeb.V1.KnownHostControllerTest do
  use ArchitectWeb.ConnCase

  alias Architect.KnownHosts
  alias Architect.KnownHosts.KnownHost

  @create_attrs %{
    comment: "some comment",
    entry: "some entry",
    fingerprint_md5: "some fingerprint_md5",
    fingerprint_sha256: "some fingerprint_sha256",
    hosts: "some hosts"
  }
  @update_attrs %{
    comment: "some updated comment",
    entry: "some updated entry",
    fingerprint_md5: "some updated fingerprint_md5",
    fingerprint_sha256: "some updated fingerprint_sha256",
    hosts: "some updated hosts"
  }
  @invalid_attrs %{
    comment: nil,
    entry: nil,
    fingerprint_md5: nil,
    fingerprint_sha256: nil,
    hosts: nil
  }

  def fixture(:known_host) do
    {:ok, known_host} = KnownHosts.create_known_host(@create_attrs)
    known_host
  end

  setup %{conn: conn} do
    {:ok, conn: put_req_header(conn, "accept", "application/json")}
  end

  describe "index" do
    test "lists all known_hosts", %{conn: conn} do
      conn = get(conn, "/v1/ssh/known-hosts")
      assert json_response(conn, 200)["data"] == []
    end
  end

  describe "create known_host" do
    test "renders known_host when data is valid", %{conn: conn} do
      conn = post(conn, "/v1/ssh/known-hosts", known_host: @create_attrs)
      assert %{"id" => id} = json_response(conn, 201)["data"]

      conn = get(conn, V1Routes.known_host_path(conn, :show, id))

      assert %{
               "id" => id,
               "comment" => "some comment",
               "entry" => "some entry",
               "fingerprint_md5" => "some fingerprint_md5",
               "fingerprint_sha256" => "some fingerprint_sha256",
               "hosts" => "some hosts"
             } = json_response(conn, 200)["data"]
    end

    test "renders errors when data is invalid", %{conn: conn} do
      conn = post(conn, "/v1/ssh/known-hosts", known_host: @invalid_attrs)
      assert json_response(conn, 422)["errors"] != %{}
    end
  end

  describe "update known_host" do
    setup [:create_known_host]

    test "renders known_host when data is valid", %{
      conn: conn,
      known_host: %KnownHost{id: id} = known_host
    } do
      conn =
        put(conn, V1Routes.known_host_path(conn, :update, known_host), known_host: @update_attrs)

      assert %{"id" => ^id} = json_response(conn, 200)["data"]

      conn = get(conn, V1Routes.known_host_path(conn, :show, id))

      assert %{
               "id" => id,
               "comment" => "some updated comment",
               "entry" => "some updated entry",
               "fingerprint_md5" => "some updated fingerprint_md5",
               "fingerprint_sha256" => "some updated fingerprint_sha256",
               "hosts" => "some updated hosts"
             } = json_response(conn, 200)["data"]
    end

    test "renders errors when data is invalid", %{conn: conn, known_host: known_host} do
      conn =
        put(conn, V1Routes.known_host_path(conn, :update, known_host), known_host: @invalid_attrs)

      assert json_response(conn, 422)["errors"] != %{}
    end
  end

  describe "delete known_host" do
    setup [:create_known_host]

    test "deletes chosen known_host", %{conn: conn, known_host: known_host} do
      conn = delete(conn, V1Routes.known_host_path(conn, :delete, known_host))
      assert response(conn, 204)

      assert_error_sent 404, fn ->
        get(conn, V1Routes.known_host_path(conn, :show, known_host))
      end
    end
  end

  defp create_known_host(_) do
    known_host = fixture(:known_host)
    {:ok, known_host: known_host}
  end
end
