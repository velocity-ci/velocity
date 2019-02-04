defmodule ArchitectWeb.Schema.EventsTypes do
  use Absinthe.Schema.Notation
  use Absinthe.Relay.Schema.Notation, :modern

  node object(:event) do
    field(:id, non_null(:string))
    field(:type, non_null(:string))

    field(:known_host, :known_host) do
      resolve(fn parent, args, res ->
        {:ok, parent.known_host}
      end)
    end

    field(:project, :project) do
      resolve(fn parent, args, res ->
        {:ok, parent.project}
      end)
    end
  end
end
