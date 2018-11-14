module Icon exposing (Options, SizeUnit(..), bell, defaultOptions, logOut, plus, plusCircle)

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



-- Icons


logOut : Options -> Element msg
logOut =
    featherIcon FeatherIcons.logOut


bell : Options -> Element msg
bell =
    featherIcon FeatherIcons.bell


plusCircle : Options -> Element msg
plusCircle =
    featherIcon FeatherIcons.plusCircle


plus : Options -> Element msg
plus =
    featherIcon FeatherIcons.plus



-- Private


featherIcon : FeatherIcons.Icon -> Options -> Element msg
featherIcon icon { size, strokeWidth, sizeUnit } =
    icon
        |> FeatherIcons.withSize size
        |> FeatherIcons.withStrokeWidth strokeWidth
        |> FeatherIcons.withSizeUnit (sizeUnitToString sizeUnit)
        |> FeatherIcons.toHtml []
        |> Element.html
