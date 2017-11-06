module Page.Project.Overview exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Project as Project exposing (Project)
import Page.Helpers exposing (formatDateTime)


-- VIEW --


view : Project -> Html msg
view project =
    div [ class "card" ]
        [ div [ class "card-body" ]
            [ dl [ style [ ( "margin-bottom", "0" ) ] ]
                [ dt [] [ text "Repository" ]
                , dd [] [ text project.repository ]
                , dt [] [ text "Last update" ]
                , dd [] [ text (formatDateTime project.updatedAt) ]
                ]
            ]
        ]
