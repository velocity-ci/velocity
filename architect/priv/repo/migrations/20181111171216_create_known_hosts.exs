defmodule Architect.Repo.Migrations.CreateKnownHosts do
  use Ecto.Migration

  def change do
    create table(:known_hosts, primary_key: false) do
      add :id, :uuid, primary_key: true
      add :entry, :string
      add :hosts, {:array, :string}
      add :comment, :string
      add :fingerprint_sha256, :string
      add :fingerprint_md5, :string

      timestamps()
    end
  end
end
