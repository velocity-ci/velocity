module Request.Errors
    exposing
        ( Error(..)
        , HttpError
        , HandledError(..)
        , handleError
        , withDefaultError
        , mapUnhandledError
        )

import Http


type HandledError
    = Unauthorized


type Error unhandled
    = HandledError HandledError
    | UnhandledError unhandled


type alias HttpError =
    Error Http.Error


handleError : Http.Error -> Error Http.Error
handleError err =
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