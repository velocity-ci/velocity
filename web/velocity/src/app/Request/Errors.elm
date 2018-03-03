module Request.Errors exposing (handle, Error(..))

import Http
import Page.Errored exposing (PageLoadError)


type Error
    = PageLoadError PageLoadError
    | UnauthorizedError


handle : Http.Error -> PageLoadError -> Error
handle err pageLoadError =
    let
        defaultError =
            PageLoadError pageLoadError
    in
        case err of
            Http.BadStatus { status } ->
                if status.code == 401 then
                    UnauthorizedError
                else
                    defaultError

            _ ->
                defaultError
