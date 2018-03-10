module Request.Errors
    exposing
        ( Error(..)
        , HttpError
        , HandledError(..)
        , handleHttpError
        , withDefaultError
        , mapUnhandledError
        )

import Request.Channel as Channel
import Http


type HandledError
    = Unauthorized


type Error unhandled
    = HandledError HandledError
    | UnhandledError unhandled


type alias HttpError =
    Error Http.Error


handleHttpError : Http.Error -> Error Http.Error
handleHttpError err =
    let
        unhandled =
            UnhandledError err
    in
        case err of
            Http.BadStatus { status } ->
                if status.code == 401 then
                    HandledError Unauthorized
                else
                    unhandled

            _ ->
                unhandled


handleChannelError : Channel.Error -> Error Channel.Error
handleChannelError err =
    case err of
        Channel.AccessDenied ->
            HandledError Unauthorized


mapUnhandledError : (Http.Error -> defaultError) -> Error Http.Error -> Error defaultError
mapUnhandledError f err =
    case err of
        UnhandledError httpError ->
            UnhandledError (f httpError)

        HandledError handled ->
            HandledError handled


withDefaultError : defaultError -> Error Http.Error -> Error defaultError
withDefaultError defaultError =
    mapUnhandledError (always defaultError)
