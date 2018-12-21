defmodule Architect.Builders.Builder do
  @moduledoc """

  """
  defstruct([
    :token,
    :state,
    :created_at,
    :updated_at
  ])

  @typedoc """
  """
  @type t :: %Architect.Builders.Builder{
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
