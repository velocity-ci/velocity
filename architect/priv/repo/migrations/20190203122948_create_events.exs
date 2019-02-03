defmodule Architect.Repo.Migrations.CreateEvents do
  use Ecto.Migration
  alias Architect.Events.Event
  alias Event.TypeEnum

  def change do

    TypeEnum.create_type()

    create table(:events, primary_key: false) do

      add(:id, :uuid, primary_key: true)
      add(:type, TypeEnum.type())

      add(:project_id, references(:projects, type: :uuid), null: true)
      add(:known_host_id, references(:known_hosts, type: :uuid), null: true)

      add(:user_id, references(:users, type: :uuid))

      add(:metadata, :map)

      timestamps()
    end

  end
end
