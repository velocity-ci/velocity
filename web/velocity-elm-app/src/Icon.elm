module Icon exposing
    ( Options
    , SizeUnit(..)
    , arrowLeft
    , arrowRight
    , bell
    , check
    , code
    , defaultOptions
    , edit
    , edit2
    , externalLink
    , fullSizeOptions
    , gitPullRequest
    , github
    , gitlab
    , link
    , link2
    , logOut
    , plus
    , plusCircle
    , settings
    , uploadCloud
    , x
    , xSquare
    )

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
    { size = 16
    , strokeWidth = 1
    , sizeUnit = Pixels
    }


fullSizeOptions : Options
fullSizeOptions =
    { size = 100
    , strokeWidth = 1
    , sizeUnit = Percentage
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


arrowRight : Options -> Element msg
arrowRight =
    featherIcon FeatherIcons.arrowRight


arrowLeft : Options -> Element msg
arrowLeft =
    featherIcon FeatherIcons.arrowLeft


x : Options -> Element msg
x =
    featherIcon FeatherIcons.x


github : Options -> Element msg
github =
    featherIcon FeatherIcons.github


gitlab : Options -> Element msg
gitlab =
    featherIcon FeatherIcons.gitlab


gitPullRequest : Options -> Element msg
gitPullRequest =
    featherIcon FeatherIcons.gitPullRequest


check : Options -> Element msg
check =
    featherIcon FeatherIcons.check


settings : Options -> Element msg
settings =
    featherIcon FeatherIcons.settings


link : Options -> Element msg
link =
    featherIcon FeatherIcons.link


link2 : Options -> Element msg
link2 =
    featherIcon FeatherIcons.link2


externalLink : Options -> Element msg
externalLink =
    featherIcon FeatherIcons.externalLink


edit : Options -> Element msg
edit =
    featherIcon FeatherIcons.edit


edit2 : Options -> Element msg
edit2 =
    featherIcon FeatherIcons.edit2


xSquare : Options -> Element msg
xSquare =
    featherIcon FeatherIcons.xSquare


uploadCloud : Options -> Element msg
uploadCloud =
    featherIcon FeatherIcons.uploadCloud


code : Options -> Element msg
code =
    featherIcon FeatherIcons.code



-- Private


featherIcon : FeatherIcons.Icon -> Options -> Element msg
featherIcon icon { size, strokeWidth, sizeUnit } =
    icon
        |> FeatherIcons.withSize size
        |> FeatherIcons.withStrokeWidth strokeWidth
        |> FeatherIcons.withSizeUnit (sizeUnitToString sizeUnit)
        |> FeatherIcons.toHtml []
        |> Element.html
