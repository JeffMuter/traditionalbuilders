-- +goose Up
CREATE TABLE IF NOT EXISTS zip_codes (
    zip   TEXT PRIMARY KEY,
    lat   REAL NOT NULL,
    lng   REAL NOT NULL,
    city  TEXT NOT NULL,
    state TEXT NOT NULL
);

-- Seed data: representative US zip codes used by seed professionals + surrounding areas
-- Source: GeoNames public domain data (geonames.org/postal-codes)
-- For production, load the full GeoNames US.zip CSV (~41k rows)

INSERT OR IGNORE INTO zip_codes (zip, lat, lng, city, state) VALUES
-- Orlando, FL area
('32801', 28.5383, -81.3792, 'Orlando', 'FL'),
('32803', 28.5572, -81.3455, 'Orlando', 'FL'),
('32804', 28.5688, -81.3955, 'Orlando', 'FL'),
('32806', 28.5122, -81.3638, 'Orlando', 'FL'),
('32812', 28.4740, -81.3455, 'Orlando', 'FL'),
('32819', 28.4500, -81.4592, 'Orlando', 'FL'),
('32836', 28.4200, -81.4978, 'Orlando', 'FL'),
('32837', 28.3889, -81.3898, 'Orlando', 'FL'),
('32789', 28.5975, -81.3511, 'Winter Park', 'FL'),
('34741', 28.2930, -81.4067, 'Kissimmee', 'FL'),
-- Charleston, SC area
('29401', 32.7765, -79.9311, 'Charleston', 'SC'),
('29403', 32.7710, -79.9481, 'Charleston', 'SC'),
('29405', 32.8083, -79.9796, 'North Charleston', 'SC'),
('29407', 32.7537, -79.9930, 'Charleston', 'SC'),
('29412', 32.7178, -79.9548, 'James Island', 'SC'),
('29414', 32.8068, -80.0585, 'West Ashley', 'SC'),
('29455', 32.6853, -80.1151, 'Johns Island', 'SC'),
('29466', 32.8878, -79.8785, 'Mount Pleasant', 'SC'),
('29483', 33.0060, -80.0585, 'Summerville', 'SC'),
-- Savannah, GA area
('31401', 32.0820, -81.0911, 'Savannah', 'GA'),
('31404', 32.0820, -81.0558, 'Savannah', 'GA'),
('31405', 32.0500, -81.1264, 'Savannah', 'GA'),
('31406', 32.0168, -81.0911, 'Savannah', 'GA'),
('31407', 32.0820, -81.1617, 'Savannah', 'GA'),
('31410', 32.0500, -80.9852, 'Savannah', 'GA'),
('31419', 32.0168, -81.1264, 'Savannah', 'GA'),
('31322', 32.1442, -81.2570, 'Pooler', 'GA'),
-- New Orleans, LA area
('70115', 29.9288, -90.0930, 'New Orleans', 'LA'),
('70116', 29.9602, -90.0607, 'New Orleans', 'LA'),
('70117', 29.9512, -90.0443, 'New Orleans', 'LA'),
('70118', 29.9397, -90.1130, 'New Orleans', 'LA'),
('70119', 29.9714, -90.0930, 'New Orleans', 'LA'),
('70122', 30.0025, -90.0443, 'New Orleans', 'LA'),
('70124', 30.0025, -90.1130, 'Lakeview', 'LA'),
('70126', 29.9714, -90.0443, 'New Orleans', 'LA'),
('70130', 29.9397, -90.0757, 'New Orleans', 'LA'),
('70002', 29.9990, -90.1627, 'Metairie', 'LA'),
('70005', 30.0033, -90.1627, 'Metairie', 'LA'),
-- Philadelphia, PA area
('19103', 39.9527, -75.1652, 'Philadelphia', 'PA'),
('19102', 39.9527, -75.1660, 'Philadelphia', 'PA'),
('19104', 39.9543, -75.1980, 'Philadelphia', 'PA'),
('19106', 39.9497, -75.1430, 'Philadelphia', 'PA'),
('19107', 39.9497, -75.1652, 'Philadelphia', 'PA'),
('19118', 40.0752, -75.2060, 'Philadelphia', 'PA'),
('19119', 40.0623, -75.1817, 'Philadelphia', 'PA'),
('19128', 40.0398, -75.2272, 'Philadelphia', 'PA'),
('19143', 39.9435, -75.2272, 'Philadelphia', 'PA'),
('19146', 39.9351, -75.1817, 'Philadelphia', 'PA'),
('08002', 39.9321, -75.0368, 'Cherry Hill', 'NJ'),
('19053', 40.1369, -74.9945, 'Feasterville', 'PA'),
-- Nashville, TN area
('37205', 36.0922, -86.8641, 'Nashville', 'TN'),
('37201', 36.1614, -86.7775, 'Nashville', 'TN'),
('37203', 36.1432, -86.7958, 'Nashville', 'TN'),
('37204', 36.1068, -86.7775, 'Nashville', 'TN'),
('37206', 36.1697, -86.7428, 'Nashville', 'TN'),
('37207', 36.2236, -86.7775, 'Nashville', 'TN'),
('37209', 36.1432, -86.8641, 'Nashville', 'TN'),
('37211', 36.0740, -86.7428, 'Nashville', 'TN'),
('37215', 36.0740, -86.8021, 'Nashville', 'TN'),
('37027', 36.0237, -86.8321, 'Brentwood', 'TN'),
('37067', 35.9257, -86.8321, 'Franklin', 'TN'),
-- Richmond, VA area
('23220', 37.5407, -77.4360, 'Richmond', 'VA'),
('23221', 37.5551, -77.4840, 'Richmond', 'VA'),
('23222', 37.5878, -77.4320, 'Richmond', 'VA'),
('23223', 37.5551, -77.3880, 'Richmond', 'VA'),
('23224', 37.5100, -77.4360, 'Richmond', 'VA'),
('23225', 37.5100, -77.5320, 'Richmond', 'VA'),
('23226', 37.5551, -77.5320, 'Richmond', 'VA'),
('23228', 37.6068, -77.5320, 'Henrico', 'VA'),
('23229', 37.5878, -77.5320, 'Henrico', 'VA'),
('23230', 37.5878, -77.4840, 'Richmond', 'VA'),
('23234', 37.4727, -77.4360, 'Richmond', 'VA'),
-- Annapolis, MD area
('21401', 38.9784, -76.4922, 'Annapolis', 'MD'),
('21403', 38.9506, -76.4748, 'Annapolis', 'MD'),
('21405', 38.9784, -76.5397, 'Annapolis', 'MD'),
('21409', 39.0063, -76.4748, 'Arnold', 'MD'),
('21032', 39.0518, -76.5872, 'Crownsville', 'MD'),
('21037', 38.9230, -76.5397, 'Edgewater', 'MD'),
('21060', 39.0063, -76.5872, 'Glen Burnie', 'MD'),
('21061', 39.1075, -76.5872, 'Glen Burnie', 'MD'),
('21108', 39.0518, -76.5397, 'Millersville', 'MD'),
('20705', 39.0063, -76.8795, 'Beltsville', 'MD'),
('21225', 39.2084, -76.6178, 'Brooklyn', 'MD'),
-- New York, NY area
('10001', 40.7484, -73.9967, 'New York', 'NY'),
('10002', 40.7157, -73.9863, 'New York', 'NY'),
('10011', 40.7421, -74.0021, 'New York', 'NY'),
('10022', 40.7587, -73.9682, 'New York', 'NY'),
('11201', 40.6920, -73.9900, 'Brooklyn', 'NY'),
('11211', 40.7126, -73.9535, 'Brooklyn', 'NY'),
-- Los Angeles, CA area
('90001', 33.9731, -118.2479, 'Los Angeles', 'CA'),
('90036', 34.0752, -118.3630, 'Los Angeles', 'CA'),
('90210', 34.0901, -118.4065, 'Beverly Hills', 'CA'),
('94102', 37.7789, -122.4194, 'San Francisco', 'CA'),
('94110', 37.7479, -122.4159, 'San Francisco', 'CA'),
-- Chicago, IL area
('60601', 41.8858, -87.6181, 'Chicago', 'IL'),
('60602', 41.8827, -87.6295, 'Chicago', 'IL'),
('60614', 41.9234, -87.6484, 'Chicago', 'IL'),
-- Houston, TX area
('77001', 29.7530, -95.3677, 'Houston', 'TX'),
('77019', 29.7612, -95.4177, 'Houston', 'TX'),
('78201', 29.4241, -98.4936, 'San Antonio', 'TX'),
-- Phoenix, AZ area
('85001', 33.4484, -112.0740, 'Phoenix', 'AZ'),
('85004', 33.4484, -112.0518, 'Phoenix', 'AZ'),
-- Seattle, WA area
('98101', 47.6062, -122.3321, 'Seattle', 'WA'),
('98103', 47.6796, -122.3500, 'Seattle', 'WA'),
-- Atlanta, GA area
('30303', 33.7490, -84.3880, 'Atlanta', 'GA'),
('30308', 33.7720, -84.3680, 'Atlanta', 'GA'),
-- Boston, MA area
('02116', 42.3489, -71.0812, 'Boston', 'MA'),
('02134', 42.3536, -71.1337, 'Boston', 'MA'),
-- Denver, CO area
('80202', 39.7501, -104.9997, 'Denver', 'CO'),
-- Miami, FL area
('33139', 25.7814, -80.1296, 'Miami Beach', 'FL'),
('33130', 25.7617, -80.2097, 'Miami', 'FL'),
-- Charlotte, NC area
('28202', 35.2320, -80.8431, 'Charlotte', 'NC'),
-- Columbus, OH area
('43201', 39.9990, -82.9988, 'Columbus', 'OH');

-- +goose Down
DROP TABLE IF EXISTS zip_codes;
