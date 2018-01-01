module Page.Project.Commit.Route exposing (Route(..), routeToPieces, route)

import UrlParser as Url exposing (parseHash, s, (</>), (<?>), string, stringParam, intParam, oneOf, Parser)
import Data.Task as ProjectTask
import Data.Build as Build
import Util exposing ((=>))


type Route
    = Overview
    | Task ProjectTask.Name (Maybe String)


route : Parser (Route -> b) b
route =
    oneOf
        [ Url.map Overview (s "overview")
        , Url.map Task (s "tasks" </> ProjectTask.nameParser <?> (Build.idQueryParser "tab"))
        ]



-- PUBLIC HELPERS --


routeToPieces : Route -> ( List String, List ( String, String ) )
routeToPieces page =
    case page of
        Overview ->
            [ "overview" ] => []

        Task name maybeBuildId ->
            let
                queryParams =
                    maybeBuildId
                        |> Maybe.map (\id -> [ "tab" => id ])
                        |> Maybe.withDefault []
            in
                [ "tasks", ProjectTask.nameToString name ] => queryParams
