module Page.Home.ActivePanel exposing (ActivePanel(..), queryParser, toQueryParams)

import Json.Decode as Decode exposing (Decoder)
import Url.Builder
import Url.Parser.Query as QueryParser


type ActivePanel
    = NewProjectForm
    | ConfigureProjectForm String


toQueryParams : Maybe ActivePanel -> List Url.Builder.QueryParameter
toQueryParams maybeActivePanel =
    case maybeActivePanel of
        Just NewProjectForm ->
            [ Url.Builder.string "active-panel" "new" ]

        Just (ConfigureProjectForm repository) ->
            [ Url.Builder.string "active-panel" "configure"
            , Url.Builder.string "repository" repository
            ]

        Nothing ->
            []


queryParser : QueryParser.Parser (Maybe ActivePanel)
queryParser =
    QueryParser.map2
        (\maybeActivePanel maybeRepository ->
            case ( maybeActivePanel, maybeRepository ) of
                ( Just "configure", Just repository ) ->
                    Just (ConfigureProjectForm repository)

                ( Just "new", _ ) ->
                    Just NewProjectForm

                _ ->
                    Nothing
        )
        (QueryParser.string "active-panel")
        (QueryParser.string "repository")
