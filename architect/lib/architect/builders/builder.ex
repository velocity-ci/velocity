defmodule Architect.Builders.Builder do
  @moduledoc """

  """
  defstruct([
    :id,
    :token,
    :state,
    :created_at,
    :updated_at
  ])

  @typedoc """
  """
  @type t :: %Architect.Builders.Builder{
          id: String.t(),
          token: String.t(),
          state: String.t(),
          created_at: Time.t(),
          updated_at: Time.t()
        }

  def state_ready, do: "ready"
  def state_busy, do: "busy"
  def state_error, do: "error"
  def state_disconnected, do: "disconnected"
end
