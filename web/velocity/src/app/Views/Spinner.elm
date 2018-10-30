module Views.Spinner exposing (spinner)

import Html exposing (Attribute, Html, i)
import Html.Attributes exposing (class, style)
import Util exposing ((=>))


spinner : Html msg
spinner =
    i [ class "fa fa-cog fa-spin" ] []
