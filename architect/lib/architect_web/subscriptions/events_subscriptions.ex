defmodule ArchitectWeb.Subscriptions.EventsSubscriptions do
  use Absinthe.Schema.Notation
  alias Ecto.Changeset
  alias Architect.Events.Event

  object :events_subscriptions do
    field :event_added, non_null(:event) do
      config(fn _args, _info ->
        {:ok, topic: "all"}
      end)
    end
  end
end
