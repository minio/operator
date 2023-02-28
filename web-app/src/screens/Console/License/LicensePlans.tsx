//  This file is part of MinIO Console Server
//  Copyright (c) 2022 MinIO, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

import React, { Fragment, useEffect, useState } from "react";
import clsx from "clsx";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import {
  AGPLV3Logo,
  Button,
  ConsoleEnterprise,
  ConsoleStandard,
  LicenseDocIcon,
} from "mds";
import { useTheme } from "@mui/material/styles";
import { SubnetInfo } from "./types";
import { Box } from "@mui/material";
import useMediaQuery from "@mui/material/useMediaQuery";
import {
  COMMUNITY_PLAN_FEATURES,
  ENTERPRISE_PLAN_FEATURES,
  FEATURE_ITEMS,
  getRenderValue,
  LICENSE_PLANS,
  PAID_PLANS,
  STANDARD_PLAN_FEATURES,
} from "./utils";

interface IRegisterStatus {
  activateProductModal: any;
  closeModalAndFetchLicenseInfo: any;
  licenseInfo: SubnetInfo | undefined;
  operatorMode: boolean;
  currentPlanID: number;
  setActivateProductModal: any;
}

const PlanHeader = ({
  isActive,
  isXsViewActive,
  title,
  onClick,
  children,
}: {
  isActive: boolean;
  isXsViewActive: boolean;
  title: string;
  price?: string;
  onClick: any;
  children: any;
}) => {
  const plan = title.toLowerCase();
  return (
    <Box
      className={clsx({
        "plan-header": true,
        active: isActive,
        [`xs-active`]: isXsViewActive,
      })}
      onClick={() => {
        onClick && onClick(plan);
      }}
      sx={{
        display: "flex",
        alignItems: "flex-start",
        justifyContent: "center",
        flexFlow: "column",
        borderLeft: "1px solid #eaeaea",
        borderBottom: "0px !important",
        "& .plan-header": {
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          flexFlow: "column",
        },

        "& .title-block": {
          display: "flex",
          alignItems: "center",
          flexFlow: "column",
          width: "100%",
          "& .title-main": {
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            flex: 1,
          },
          "& .iconContainer": {
            "& .min-icon": {
              minWidth: 140,
              width: "100%",
              maxHeight: 55,
              height: "100%",
            },
          },
        },

        "& .open-source": {
          fontSize: "14px",
          display: "flex",
          marginBottom: "5px",
          alignItems: "center",
          "& .min-icon": {
            marginRight: "8px",
            height: "12px",
            width: "12px",
          },
        },

        "& .cur-plan-text": {
          fontSize: "12px",
          textTransform: "uppercase",
        },

        "@media (max-width: 600px)": {
          cursor: "pointer",
          "& .title-block": {
            "& .title": {
              fontSize: "14px",
              fontWeight: 600,
            },
          },
        },

        "&.active, &.active.xs-active": {
          color: "#ffffff",
          position: "relative",

          "& .min-icon": {
            fill: "#ffffff",
          },

          "&:before": {
            content: "' '",
            position: "absolute",
            width: "100%",
            height: "18px",
            backgroundColor: "#2781B0",
            display: "block",
            top: -16,
          },
          "& .iconContainer": {
            "& .min-icon": {
              marginTop: "-12px",
            },
          },
        },
        "&.active": {
          background: "#2781B0",
          color: "#ffffff",
        },
        "&.xs-active": {
          background: "#eaeaea",
        },
      }}
    >
      {children}
    </Box>
  );
};

const FeatureTitleRowCmp = (props: { featureLabel: any }) => {
  return (
    <Box className="feature-title">
      <Box className="feature-title-info">
        <div className="xs-only">{props.featureLabel} </div>
      </Box>
    </Box>
  );
};

const PricingFeatureItem = (props: {
  featureLabel: any;
  label?: any;
  detail?: any;
  xsLabel?: string;
  style?: any;
}) => {
  return (
    <Box className="feature-item" style={props.style}>
      <Box className="feature-item-info">
        <div className="xs-only">
          {getRenderValue(props.featureLabel || "")}
        </div>
        <Box className="plan-feature">
          <div>{getRenderValue(props.label || "")}</div>
          {getRenderValue(props.detail)}

          <div className="xs-only">{props.xsLabel} </div>
        </Box>
      </Box>
    </Box>
  );
};

const LicensePlans = ({ licenseInfo, operatorMode }: IRegisterStatus) => {
  const theme = useTheme();
  const isSmallScreen = useMediaQuery(theme.breakpoints.down("sm"));

  let currentPlan = !licenseInfo
    ? "community"
    : licenseInfo?.plan?.toLowerCase();

  const isCommunityPlan = currentPlan === LICENSE_PLANS.COMMUNITY;
  const isStandardPlan = currentPlan === LICENSE_PLANS.STANDARD;
  const isEnterprisePlan = currentPlan === LICENSE_PLANS.ENTERPRISE;

  const isPaidPlan = PAID_PLANS.includes(currentPlan);

  /*In smaller screen use tabbed view to show features*/
  const [xsPlanView, setXsPlanView] = useState("");
  let isXsViewCommunity = xsPlanView === LICENSE_PLANS.COMMUNITY;
  let isXsViewStandard = xsPlanView === LICENSE_PLANS.STANDARD;
  let isXsViewEnterprise = xsPlanView === LICENSE_PLANS.ENTERPRISE;

  const getCommunityPlanHeader = () => {
    return (
      <PlanHeader
        key={"community-header"}
        isActive={isCommunityPlan}
        isXsViewActive={isXsViewCommunity}
        title={"community"}
        onClick={isSmallScreen ? onPlanClick : null}
      >
        <Box className="title-block">
          <Box className="title-main">
            <div className="iconContainer">
              <AGPLV3Logo style={{ width: 117 }} />
            </div>
          </Box>
        </Box>
      </PlanHeader>
    );
  };

  const getStandardPlanHeader = () => {
    return (
      <PlanHeader
        key={"standard-header"}
        isActive={isStandardPlan}
        isXsViewActive={isXsViewStandard}
        title={"Standard"}
        onClick={isSmallScreen ? onPlanClick : null}
      >
        <Box className="title-block">
          <Box className="title-main">
            <div className="iconContainer">
              <ConsoleStandard />
            </div>
          </Box>
        </Box>
      </PlanHeader>
    );
  };

  const getEnterpriseHeader = () => {
    return (
      <PlanHeader
        key={"enterprise-header"}
        isActive={isEnterprisePlan}
        isXsViewActive={isXsViewEnterprise}
        title={"Enterprise"}
        onClick={isSmallScreen ? onPlanClick : null}
      >
        <Box className="title-block">
          <Box className="title-main">
            <div className="iconContainer">
              <ConsoleEnterprise />
            </div>
          </Box>
        </Box>
      </PlanHeader>
    );
  };

  const getButton = (
    link: string,
    btnText: string,
    variant: any,
    plan: string
  ) => {
    let linkToNav =
      currentPlan !== "community" ? "https://subnet.min.io" : link;
    return (
      <Button
        id={`license-action-${link}`}
        variant={variant}
        style={{
          marginTop: "12px",
          width: "80%",
        }}
        disabled={
          currentPlan !== LICENSE_PLANS.COMMUNITY && currentPlan !== plan
        }
        onClick={(e) => {
          e.preventDefault();

          window.open(
            `${linkToNav}?ref=${operatorMode ? "op" : "con"}`,
            "_blank"
          );
        }}
        label={btnText}
      />
    );
  };

  const onPlanClick = (plan: string) => {
    setXsPlanView(plan);
  };

  useEffect(() => {
    if (isSmallScreen) {
      setXsPlanView(currentPlan || "community");
    } else {
      setXsPlanView("");
    }
  }, [isSmallScreen, currentPlan]);

  const linkTracker = `?ref=${operatorMode ? "op" : "con"}`;

  const featureList = FEATURE_ITEMS;
  return (
    <Fragment>
      <Box
        sx={{
          border: "1px solid #eaeaea",
          borderTop: "0px",
          marginBottom: "45px",
          "&::-webkit-scrollbar": {
            width: "5px",
            height: "5px",
          },
          "&::-webkit-scrollbar-track": {
            background: "#F0F0F0",
            borderRadius: 0,
            boxShadow: "inset 0px 0px 0px 0px #F0F0F0",
          },
          "&::-webkit-scrollbar-thumb": {
            background: "#777474",
            borderRadius: 0,
          },
          "&::-webkit-scrollbar-thumb:hover": {
            background: "#5A6375",
          },
        }}
      >
        <Box
          className={"title-blue-bar"}
          sx={{
            height: "8px",
            borderBottom: "8px solid rgb(6 48 83)",
          }}
        />
        <Box
          className={isPaidPlan ? "paid-plans-only" : ""}
          sx={{
            display: "grid",

            margin: "0 1.5rem 0 1.5rem",

            gridTemplateColumns: {
              sm: "1fr 1fr 1fr 1fr",
              xs: "1fr 1fr 1fr",
            },

            "&.paid-plans-only": {
              display: "grid",
              gridTemplateColumns: "1fr 1fr 1fr",
            },

            "& .features-col": {
              flex: 1,
              minWidth: "260px",

              "@media (max-width: 600px)": {
                display: "none",
              },
            },

            "& .xs-only": {
              display: "none",
            },

            "& .button-box": {
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              padding: "5px 0px 25px 0px",
              borderLeft: "1px solid #eaeaea",
            },
            "& .plan-header": {
              height: "99px",
              borderBottom: "1px solid #eaeaea",
            },
            "& .feature-title": {
              height: "25px",
              paddingLeft: "26px",
              fontSize: "14px",

              background: "#E5E5E5",

              "@media (max-width: 600px)": {
                "& .feature-title-info .xs-only": {
                  display: "block",
                },
              },
            },
            "& .feature-name": {
              minHeight: "60px",
              padding: "5px",
              borderBottom: "1px solid #eaeaea",
              display: "flex",
              alignItems: "center",
              paddingLeft: "26px",
              fontSize: "14px",
            },
            "& .feature-item": {
              display: "flex",
              flexFlow: "column",
              alignItems: "center",
              justifyContent: "center",
              minHeight: "60px",
              padding: "0 15px 0 15px",
              borderBottom: "1px solid #eaeaea",
              borderLeft: " 1px solid #eaeaea",
              fontSize: "14px",
              "& .link-text": {
                color: "#2781B0",
                cursor: "pointer",
                textDecoration: "underline",
              },

              "&.icon-yes": {
                width: "15px",
                height: "15px",
              },
            },

            "& .feature-item-info": {
              flex: 1,
              display: "flex",
              flexFlow: "column",
              alignItems: "center",
              justifyContent: "space-around",
              textAlign: "center",

              "@media (max-width: 600px)": {
                justifyContent: "space-evenly",
                width: "100%",
                "& .xs-only": {
                  display: "block",
                },
                "& .plan-feature": {
                  textAlign: "center",
                  paddingRight: "10px",
                },
              },
            },

            "& .plan-col": {
              minWidth: "260px",
              flex: 1,
            },

            "& .active-plan-col": {
              background: "#FDFDFD 0% 0% no-repeat padding-box",
              boxShadow: " 0px 3px 20px #00000038",

              "& .plan-header": {
                backgroundColor: "#2781B0",
              },

              "& .feature-title": {
                background: "#F7F7F7",
              },
            },
          }}
        >
          <Box className="features-col">
            {featureList.map((fi) => {
              const featureTitleRow = fi.featureTitleRow;
              const isHeader = fi.isHeader;

              if (isHeader) {
                if (isPaidPlan) {
                  return (
                    <Box
                      key={`plan-header-${fi.desc}`}
                      className="plan-header"
                      sx={{
                        fontSize: "14px",
                        paddingLeft: "26px",
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "flex-start",
                        borderBottom: "0px !important",

                        "& .link-text": {
                          color: "#2781B0",
                          cursor: "pointer",
                          textDecoration: "underline",
                        },

                        "& .min-icon": {
                          marginRight: "10px",
                          color: "#2781B0",
                          fill: "#2781B0",
                        },
                      }}
                    >
                      <LicenseDocIcon />
                      <a
                        href={`https://subnet.min.io/terms-and-conditions/${currentPlan}`}
                        rel="noopener"
                        className={"link-text"}
                      >
                        View License agreement <br />
                        for the registered plan.
                      </a>
                    </Box>
                  );
                }

                return (
                  <Box
                    key={`plan-header-label-${fi.desc}`}
                    className={`plan-header`}
                    sx={{
                      fontSize: "14px",
                      paddingLeft: "26px",
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "flex-start",
                      borderBottom: "0px !important",
                    }}
                  >
                    {fi.label}
                  </Box>
                );
              }
              if (featureTitleRow) {
                return (
                  <Box
                    key={`plan-descript-${fi.desc}`}
                    className="feature-title"
                    sx={{
                      fontSize: "14px",
                      fontWeight: 600,
                      textTransform: "uppercase",
                    }}
                  >
                    <div>{getRenderValue(fi.desc)} </div>
                  </Box>
                );
              }
              return (
                <Box
                  key={`plan-feature-name-${fi.desc}`}
                  className="feature-name"
                  style={fi.style}
                >
                  <div>{getRenderValue(fi.desc)} </div>
                </Box>
              );
            })}
          </Box>
          {!isPaidPlan ? (
            <Box
              className={`plan-col ${
                isCommunityPlan ? "active-plan-col" : "non-active-plan-col"
              }`}
            >
              {COMMUNITY_PLAN_FEATURES.map((fi, idx) => {
                const featureLabel = featureList[idx].desc;
                const { featureTitleRow, isHeader } = fi;

                if (isHeader) {
                  return getCommunityPlanHeader();
                }

                if (featureTitleRow) {
                  return (
                    <FeatureTitleRowCmp
                      key={`title-row-${fi.id}`}
                      featureLabel={featureLabel}
                    />
                  );
                }

                return (
                  <PricingFeatureItem
                    key={`pricing-feature-${fi.id}`}
                    featureLabel={featureLabel}
                    label={fi.label}
                    detail={fi.detail}
                    xsLabel={fi.xsLabel}
                    style={fi.style}
                  />
                );
              })}
              <Box className="button-box">
                {getButton(
                  `https://slack.min.io${linkTracker}`,
                  "Join Slack",
                  "regular",
                  LICENSE_PLANS.COMMUNITY
                )}
              </Box>
            </Box>
          ) : null}
          <Box
            className={`plan-col ${
              isStandardPlan ? "active-plan-col" : "non-active-plan-col"
            }`}
          >
            {STANDARD_PLAN_FEATURES.map((fi, idx) => {
              const featureLabel = featureList[idx].desc;
              const featureTitleRow = fi.featureTitleRow;
              const isHeader = fi.isHeader;

              if (isHeader) {
                return getStandardPlanHeader();
              }

              if (featureTitleRow) {
                return (
                  <FeatureTitleRowCmp
                    key={`feature-title-row-${fi.id}`}
                    featureLabel={featureLabel}
                  />
                );
              }
              return (
                <PricingFeatureItem
                  key={`feature-item-${fi.id}`}
                  featureLabel={featureLabel}
                  label={fi.label}
                  detail={fi.detail}
                  xsLabel={fi.xsLabel}
                  style={fi.style}
                />
              );
            })}

            <Box className="button-box">
              {getButton(
                `https://min.io/signup${linkTracker}`,
                !PAID_PLANS.includes(currentPlan)
                  ? "Subscribe"
                  : "Login to SUBNET",
                "callAction",
                LICENSE_PLANS.STANDARD
              )}
            </Box>
          </Box>
          <Box
            className={`plan-col ${
              isEnterprisePlan ? "active-plan-col" : "non-active-plan-col"
            }`}
          >
            {ENTERPRISE_PLAN_FEATURES.map((fi, idx) => {
              const featureLabel = featureList[idx].desc;
              const { featureTitleRow, isHeader, yesIcon } = fi;

              if (isHeader) {
                return getEnterpriseHeader();
              }

              if (featureTitleRow) {
                return (
                  <FeatureTitleRowCmp
                    key={`feature-title-row2-${fi.id}`}
                    featureLabel={featureLabel}
                  />
                );
              }

              if (yesIcon) {
                return (
                  <Box className="feature-item" key={`ent-feature-yes${fi.id}`}>
                    <Box className="feature-item-info">
                      <div className="xs-only"></div>
                      <Box className="plan-feature">
                        <CheckCircleIcon />
                      </Box>
                    </Box>
                  </Box>
                );
              }
              return (
                <PricingFeatureItem
                  key={`pricing-feature-item-${fi.id}`}
                  featureLabel={featureLabel}
                  label={fi.label}
                  detail={fi.detail}
                  style={fi.style}
                />
              );
            })}
            <Box className="button-box">
              {getButton(
                `https://min.io/signup${linkTracker}`,
                !PAID_PLANS.includes(currentPlan)
                  ? "Subscribe"
                  : "Login to SUBNET",
                "callAction",
                LICENSE_PLANS.ENTERPRISE
              )}
            </Box>
          </Box>
        </Box>
      </Box>
    </Fragment>
  );
};

export default LicensePlans;
