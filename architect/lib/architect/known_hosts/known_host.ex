defmodule Architect.KnownHosts.KnownHost do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :binary_id, autogenerate: true}
  schema "known_hosts" do
    field(:entry, :string)
    field(:fingerprint_md5, :string)
    field(:fingerprint_sha256, :string)
    field(:host, :string)
    field(:user_verified, :boolean)

    timestamps()
  end

  @doc false
  def changeset(known_host, attrs) do
    known_host
    |> cast(attrs, [:entry, :hosts, :comment, :fingerprint_sha256, :fingerprint_md5])
    # TODO: parse entry instead
    |> validate_required([:entry, :hosts, :comment, :fingerprint_sha256, :fingerprint_md5])
  end

  def scan_host(host) do
    {output, exit_code} = System.cmd("#{System.cwd()}/v-ssh-keyscan", [host])

    if exit_code != 0 do
      # panic
    end

    Poison.decode!(output)
  end
end
