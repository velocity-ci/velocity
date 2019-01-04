defmodule Architect.KnownHosts.KnownHost do
  use Ecto.Schema
  import Ecto.Changeset
  alias Ecto.Changeset
  alias Architect.KnownHosts.Scanned
  require Logger

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
  def changeset(%__MODULE__{} = known_host, attrs) do
    known_host
    |> cast(attrs, [:host, :verified])
    |> validate_required([:host])
    |> unique_constraint(:host)
    |> populate()
  end

  @doc """
  Populate a KnownHost changeset by scanning the value specified at host, if changeset is valid
  """
  def populate(%Changeset{valid?: true, changes: %{host: host}} = changeset) do
    case Scanned.generate(host) do
      {:error, _} ->
        add_error(changeset, :host, "Scanning host failed")

      {:ok, scanned} ->
        populate(changeset, scanned)
    end
  end

  def populate(changeset), do: changeset

  @doc """
  Populate a KnownHost changeset with the values specified in the Scanned struct, if valid
  """
  def populate(%Changeset{valid?: true} = changeset, %Scanned{} = scanned) do
    changeset
    |> put_change(:entry, scanned.entry)
    |> put_change(:fingerprint_md5, scanned.md5)
    |> put_change(:fingerprint_sha256, scanned.sha256)
    |> put_change(:verified, false)
  end

  def populate(changeset, _scanned), do: changeset
end
