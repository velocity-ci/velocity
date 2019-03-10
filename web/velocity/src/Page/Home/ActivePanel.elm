module Page.Home.ActivePanel exposing (ActivePanel(..), queryParser, toQueryParams)

import Json.Decode as Decode exposing (Decoder)
import Url.Builder
import Url.Parser.Query as QueryParser


type ActivePanel
    = None
    | ProjectForm


toQueryParams : ActivePanel -> List Url.Builder.QueryParameter
toQueryParams activePanel =
    case activePanel of
        ProjectForm ->
            [ Url.Builder.string "active-panel" "project-form" ]

        None ->
            []


queryParser : QueryParser.Parser ActivePanel
queryParser =
    QueryParser.map
        (\activePanel ->
            case activePanel of
                Just "project-form" ->
                    ProjectForm

                _ ->
                    None
        )
        (QueryParser.string "active-panel")
