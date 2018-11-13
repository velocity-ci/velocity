module Icon exposing (Options, SizeUnit(..), defaultOptions, logOut)

import Element exposing (Element)
import FeatherIcons


type SizeUnit
    = Percentage
    | Pixels


sizeUnitToString : SizeUnit -> String
sizeUnitToString sizeUnit =
    case sizeUnit of
        Percentage ->
            "%"

        Pixels ->
            "px"


type alias Options =
    { size : Float
    , strokeWidth : Float
    , sizeUnit : SizeUnit
    }


defaultOptions : Options
defaultOptions =
    { size = 24
    , strokeWidth = 1
    , sizeUnit = Pixels
    }


logOut : Options -> Element msg
logOut { size, strokeWidth, sizeUnit } =
    FeatherIcons.logOut
        |> FeatherIcons.withSize size
        |> FeatherIcons.withStrokeWidth strokeWidth
        |> FeatherIcons.withSizeUnit (sizeUnitToString sizeUnit)
        |> FeatherIcons.toHtml []
        |> Element.html
