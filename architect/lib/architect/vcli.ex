defmodule Architect.VCLI do
  @moduledoc """
  Deals with calling the vcli binary
  """

  require Logger

  def init(), do: Application.get_env(:architect, :vcli) |> Enum.into(%{})

  def list(opts), do: cmd(opts, ["list", "--machine-readable"])

  defp cmd(%{bin: bin, timeout: timeout, log_errors: log_errors}, cmd) when is_list(cmd) do
    try do
      {out, _} =
        Task.async(fn ->
          System.cmd(bin, cmd, stderr_to_stdout: true)
        end)
        |> Task.await(timeout)

      Poison.decode!(out)
    catch
      :exit, _ ->
        log("VCLI timeout", :error, log_errors)
        {:error, :timeout}
    end
  end

  @doc false
  defp log(output, :debug, true), do: Logger.debug(output)
  defp log(output, :warn, true), do: Logger.warn(output)
  defp log(output, :error, true), do: Logger.error(output)
  defp log(_, _, _), do: nil
end
