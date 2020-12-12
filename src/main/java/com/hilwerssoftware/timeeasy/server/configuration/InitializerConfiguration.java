package com.hilwerssoftware.timeeasy.server.configuration;

import com.hilwerssoftware.timeeasy.server.initialization.ProductionDataInitializer;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class InitializerConfiguration {

    @Bean
    public ProductionDataInitializer productionDataInitializer() {
        return new ProductionDataInitializer();
    }
}
