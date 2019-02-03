defmodule Architect.Repo.Migrations.CreateKnownHostEvents do
  use Ecto.Migration
  alias Architect.KnownHosts.Event
  alias Event.TypeEnum

  def change do

    TypeEnum.create_type()

    create table(:known_host_events, primary_key: false) do

      add(:id, :uuid, primary_key: true)
      add :type, TypeEnum.type()

      add(:known_host_id, references(:known_hosts, type: :uuid))
      add(:user_id, references(:users, type: :uuid))

      add(:metadata, :map)
      timestamps()

    end

  end
end