-- +goose Up
INSERT INTO professionals (name, email, phone, specialty, location, bio, zip_code, latitude, longitude) VALUES
(
    'John Smith',
    'john.smith@classicbuilders.com',
    '(407) 555-0101',
    'Georgian Architecture',
    'Orlando, FL',
    'Award-winning architect with 20 years specializing in Georgian Revival residential and civic buildings. Member of the Institute of Classical Architecture & Art.',
    '32801',
    28.5383,
    -81.3792
),
(
    'Margaret Chen',
    'margaret.chen@colonialcraft.com',
    '(843) 555-0102',
    'Colonial Revival',
    'Charleston, SC',
    'Boutique firm focused on historically accurate Colonial Revival homes in the Lowcountry. Restoration and new construction with period-correct materials.',
    '29401',
    32.7765,
    -79.9311
),
(
    'William Hartford',
    'william.hartford@federalworks.com',
    '(912) 555-0103',
    'Federal Style Architecture',
    'Savannah, GA',
    'Specializing in Federal and Adamesque detailing. Works closely with Savannah historic preservation board on restorations and new infill construction.',
    '31401',
    32.0820,
    -81.0911
),
(
    'Elizabeth Moore',
    'elizabeth.moore@greekrevival.com',
    '(504) 555-0104',
    'Greek Revival',
    'New Orleans, LA',
    'Third-generation builder with deep roots in New Orleans architectural tradition. Expert in Greek Revival and Creole cottage styles.',
    '70115',
    29.9288,
    -90.0930
),
(
    'Robert Adams',
    'robert.adams@federalphilly.com',
    '(215) 555-0105',
    'Federal & Colonial Architecture',
    'Philadelphia, PA',
    'Licensed architect and builder specializing in 18th-century Federal and Philadelphia Colonial styles. Licensed in PA, NJ, and DE.',
    '19103',
    39.9527,
    -75.1652
),
(
    'Catherine Williams',
    'catherine.williams@classicalrevival.com',
    '(615) 555-0106',
    'Classical Revival',
    'Nashville, TN',
    'Designs and builds Classical Revival and Antebellum-inspired homes throughout Middle Tennessee. Featured in Southern Living and Architectural Digest.',
    '37205',
    36.0922,
    -86.8641
),
(
    'James Crawford',
    'james.crawford@antebellumdesign.com',
    '(804) 555-0107',
    'Antebellum & Greek Revival',
    'Richmond, VA',
    'Virginia-based firm specializing in plantation-style and Antebellum architecture. Expert in traditional timber framing and brick construction.',
    '23220',
    37.5407,
    -77.4360
),
(
    'Sarah Thompson',
    'sarah.thompson@annapolisclassic.com',
    '(410) 555-0108',
    'Colonial & Federal Architecture',
    'Annapolis, MD',
    'Award-winning designer focused on 18th-century Maryland Colonial and Federal styles. Experienced with Annapolis Historic District guidelines.',
    '21401',
    38.9784,
    -76.4922
);

-- +goose Down
DELETE FROM professionals WHERE email IN (
    'john.smith@classicbuilders.com',
    'margaret.chen@colonialcraft.com',
    'william.hartford@federalworks.com',
    'elizabeth.moore@greekrevival.com',
    'robert.adams@federalphilly.com',
    'catherine.williams@classicalrevival.com',
    'james.crawford@antebellumdesign.com',
    'sarah.thompson@annapolisclassic.com'
);
