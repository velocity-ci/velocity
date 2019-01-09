defmodule Architect.Repo.Migrations.CreateKnownHosts do
  use Ecto.Migration

  def change do
    create table(:known_hosts, primary_key: false) do
      add(:id, :uuid, primary_key: true)
      add(:entry, :text)
      add(:host, :string)
      add(:fingerprint_sha256, :string)
      add(:fingerprint_md5, :string)
      add(:verified, :boolean)

      timestamps()
    end

    create(unique_index(:known_hosts, [:host]))
  end
end
