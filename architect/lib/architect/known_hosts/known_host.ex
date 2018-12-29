defmodule Architect.KnownHosts.KnownHost do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :binary_id, autogenerate: true}
  schema "known_hosts" do
    field(:entry, :string)
    field(:host, :string)
    field(:fingerprint_md5, :string)
    field(:fingerprint_sha256, :string)
    field(:verified, :boolean, default: false)

    timestamps()
  end

  @doc false
  def changeset(known_host, attrs) do
    known_host
    |> cast(attrs, [:host])
    |> validate_required([:host])
    |> unique_constraint(:host)
  end

  def scan_host(host) do
    {output, exit_code} = System.cmd("#{System.cwd()}/v-ssh-keyscan", [host])

    if exit_code != 0 do
      # panic
    end

    Poison.decode!(output)
  end

  def populate(known_host) do
    scanned_host = scan_host(get_field(known_host, :host))

    if scanned_host["sha256Fingerprint"] == get_field(known_host, :fingerprint_sha256) do
      known_host
    else
      known_host
      |> put_change(:entry, scanned_host["entry"])
      |> put_change(:fingerprint_md5, scanned_host["md5Fingerprint"])
      |> put_change(:fingerprint_sha256, scanned_host["sha256Fingerprint"])
      |> put_change(:verified, false)
    end
  end
end
