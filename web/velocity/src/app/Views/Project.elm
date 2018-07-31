module Views.Project exposing (badge)

import Html exposing (Html)
import Html.Styled.Attributes as Attributes exposing (attribute, css, class, classList, src)
import Html.Styled as Styled exposing (..)
import Css exposing (..)
import Data.Project exposing (Project)


badge : Project -> Html.Html msg
badge project =
    case project.logo of
        Just logoHref ->
            img
                [ class "img-thumbnail img-fluid"
                , src logoHref
                ]
                []
                |> toUnstyled

        Nothing ->
            div
                [ class "badge badge-info"
                , css [ padding (px 10) ]
                ]
                [ i
                    [ attribute "aria-hidden" "true", class "fa fa-code" ]
                    []
                ]
                |> toUnstyled
