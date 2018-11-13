module Icon exposing (Options, SizeUnit(..), bell, defaultOptions, logOut)

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
logOut options =
    featherIcon FeatherIcons.logOut options


bell : Options -> Element msg
bell options =
    featherIcon FeatherIcons.bell options


featherIcon : FeatherIcons.Icon -> Options -> Element msg
featherIcon icon { size, strokeWidth, sizeUnit } =
    icon
        |> FeatherIcons.withSize size
        |> FeatherIcons.withStrokeWidth strokeWidth
        |> FeatherIcons.withSizeUnit (sizeUnitToString sizeUnit)
        |> FeatherIcons.toHtml []
        |> Element.html
