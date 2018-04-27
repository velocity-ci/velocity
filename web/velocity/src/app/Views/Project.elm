module Views.Project exposing (badge)

import Html exposing (..)
import Html.Attributes exposing (..)


badge : Html msg
badge =
    div
        [ class "badge badge-info project-badge" ]
        [ i [ attribute "aria-hidden" "true", class "fa fa-code" ] [] ]
