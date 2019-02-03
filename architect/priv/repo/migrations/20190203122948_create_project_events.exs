defmodule Architect.Repo.Migrations.CreateProjectEvents do
  use Ecto.Migration
  alias Architect.Projects.Event
  alias Event.TypeEnum

  def change do

    TypeEnum.create_type()

    create table(:project_events, primary_key: false) do

      add(:id, :uuid, primary_key: true)
      add :type, TypeEnum.type()

      add(:project_id, references(:projects, type: :uuid))
      add(:user_id, references(:users, type: :uuid))

      add(:metadata, :map)
      timestamps()

    end

  end
end
