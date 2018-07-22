module Views.Project exposing (badge)

import Html exposing (Html)
import Html.Styled.Attributes as Attributes exposing (attribute, css, class, classList)
import Html.Styled as Styled exposing (..)
import Css exposing (..)


badge : Html.Html msg
badge =
    div
        [ class "badge badge-info"
        , css [ padding (px 10) ]
        ]
        [ i
            [ attribute "aria-hidden" "true", class "fa fa-code" ]
            []
        ]
        |> toUnstyled
